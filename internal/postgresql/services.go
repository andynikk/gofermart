package postgresql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/cryptography"
	"github.com/andynikk/gofermart/internal/token"
)

func (dbc *DBConnector) NewAccount(name string, password string) (*Account, error) {
	account := NewAccount()

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectUserWithWhereTemplate, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&account.Name, &account.Password)
		if err != nil {
			return nil, err
		}
		account.ResponseStatus = constants.AnswerLoginBusy
		return account, nil
	}

	heshVal := cryptography.HeshSHA256(password, dbc.Cfg.Key)
	if _, err := conn.Exec(ctx, constants.QueryAddUserTemplate, name, heshVal); err != nil {
		return nil, err
	}
	conn.Release()

	account.Name = name
	account.Password = heshVal
	account.ResponseStatus = constants.AnswerSuccessfully

	return account, nil
}

func (dbc *DBConnector) GetAccount(name string, password string) (*Account, error) {
	account := NewAccount()

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	heshVal := cryptography.HeshSHA256(password, dbc.Cfg.Key)
	rows, err := conn.Query(ctx, constants.QuerySelectUserWithPasswordTemplate, name, heshVal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {

		_ = rows.Scan(&account.Name, &account.Password)
		account.ResponseStatus = constants.AnswerSuccessfully
		return account, nil

	}
	conn.Release()

	account.ResponseStatus = constants.AnswerInvalidLoginPassword
	return account, nil
}

func (dbc *DBConnector) NewOrder(tkn string, number int) (*Order, error) {
	order := NewOrder()

	ctx := context.Background()
	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		order.ResponseStatus = constants.AnswerUserNotAuthenticated
		return order, nil
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, "", number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {

		_ = rows.Scan(&order.User, &order.Number, &order.CreatedAt,
			&order.StartedAt, &order.FinishedAt, &order.FailedAt, &order.Status)
		if order.User == claims["user"] {
			order.ResponseStatus = constants.AnswerSuccessfully
		} else {
			order.ResponseStatus = constants.AnswerUploadedAnotherUser
		}

		return order, nil
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, constants.QueryAddOrderTemplate, claims["user"], number, time.Now()); err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}
	/////////////////////////////////////////////////////////////////

	order.ResponseStatus = constants.AnswerAccepted
	return order, nil
}

func (dbc *DBConnector) SetStartedAt(number int, tkn string) (*Order, error) {
	order := NewOrder()

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	timeNow := time.Now()
	if _, err = conn.Query(ctx, constants.QueryUpdateStartedAt, timeNow, number); err != nil {
		return nil, err
	}

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		order.Number = strconv.Itoa(number)
		order.StartedAt = timeNow
		order.ResponseStatus = constants.AnswerUserNotAuthenticated

		return order, nil
	}

	order.Number = strconv.Itoa(number)
	order.User = claims["user"].(string)
	order.StartedAt = timeNow
	order.ResponseStatus = constants.AnswerSuccessfully

	return order, nil
}

func (dbc *DBConnector) AddAccrual() {

}

func (dbc *DBConnector) TryWithdraw(tkn string, number string, sumWithdraw float64) (*Balance, error) {
	balance := NewBalance()

	balance.Number = number
	balance.Withdrawn = sumWithdraw

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		balance.ResponseStatus = constants.AnswerUserNotAuthenticated
		return balance, nil
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderBalansTemplate, claims["user"], number)
	if err != nil {
		return nil, err
	}

	var bdb BalanceDB
	if rows.Next() {

		err = rows.Scan(&bdb.Number, &bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			return nil, err
		}
		if bdb.Total < sumWithdraw {
			balance.ResponseStatus = constants.AnswerInsufficientFunds
			return balance, nil
		}
	}
	balance.BalanceDB = &bdb
	conn.Release()

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	//TODO: Добавляем спсанные баллы
	if _, err = conn.Query(ctx, constants.QueryAddAccrual, sumWithdraw, time.Now(), "MINUS", number); err != nil {
		return nil, err
	}
	rows.Close()

	conn.Release()
	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err = conn.Query(ctx, constants.QueryOrderWhereNumTemplate, claims["user"], number)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		rows.Close()
		_, err = conn.Query(ctx, constants.QueryAddOrderTemplate, claims["user"], number, time.Now())
		if err != nil {
			return nil, err
		}
	}

	balance.ResponseStatus = constants.AnswerSuccessfully
	return balance, nil
}

func (dbc *DBConnector) ListOrder(tkn string, addressAcSys string) (*OrdersDB, error) {
	ordersDB := NewOrdersDB()

	ctx := context.Background()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		ordersDB.ResponseStatus = constants.AnswerUserNotAuthenticated
		return ordersDB, nil
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryListOrderTemplate, claims["user"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ord OrderDB

		err = rows.Scan(&ord.Number, &ord.Status, &ord.Accrual, &ord.UploadedAt)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		fmt.Println()
		ss, err := GetOrder4AS(addressAcSys, ord.Number)
		if err == nil {
			ord.Status = ss.Status
			ord.Accrual = ss.Accrual
		}
		ordersDB.OrderDB = append(ordersDB.OrderDB, ord)
	}
	if len(ordersDB.OrderDB) == 0 {
		ordersDB.ResponseStatus = constants.AnswerNoContent
		return ordersDB, nil
	}

	ordersDB.ResponseStatus = constants.AnswerSuccessfully
	return ordersDB, nil
}

func (dbc *DBConnector) BalancesOrders(tkn string, addressAcSys string) (*Balances, error) {
	balances := NewBalances()

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		balances.ResponseStatus = constants.AnswerUserNotAuthenticated
		return balances, nil
	}

	//rows, err := conn.Query(ctx, constants.QueryUserBalansTemplate, claims["user"])
	rows, err := conn.Query(ctx, constants.QueryUserOrdes, claims["user"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balancesDB []BalanceDB
	for rows.Next() {
		var bdb BalanceDB

		err = rows.Scan(&bdb.Number, &bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		ss, err := GetOrder4AS(addressAcSys, bdb.Number)
		if err == nil {
			bdb.Current = ss.Accrual
		}
		balancesDB = append(balancesDB, bdb)
	}
	if len(balancesDB) == 0 {
		balances.ResponseStatus = constants.AnswerNoContent
		return balances, nil
	}

	for _, val := range balancesDB {
		balances.TotalBalanceDB.Withdrawn = balances.TotalBalanceDB.Withdrawn + val.Withdrawn
		balances.TotalBalanceDB.Current = balances.TotalBalanceDB.Current + val.Current
	}
	balances.TotalBalanceDB.Current = balances.TotalBalanceDB.Current - balances.TotalBalanceDB.Withdrawn

	if err != nil {
		return nil, err
	}

	balances.ResponseStatus = constants.AnswerSuccessfully
	return balances, nil
}

func (dbc *DBConnector) UserWithdrawal(tkn string) (*Withdraws, error) {
	withdraws := NewWithdraws()

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		withdraws.ResponseStatus = constants.AnswerUserNotAuthenticated
		return withdraws, nil
	}

	rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], "MINUS")
	if err != nil {
		withdraws.ResponseStatus = constants.AnswerInvalidFormat
		return withdraws, nil
	}
	defer rows.Close()

	for rows.Next() {
		var bdb withdrawDB

		err = rows.Scan(&bdb.Order, &bdb.DateAccrual, &bdb.Withdraw, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		withdraws.WithdrawDB = append(withdraws.WithdrawDB, bdb)
	}

	if len(withdraws.WithdrawDB) == 0 {
		withdraws.ResponseStatus = constants.AnswerNoContent
		return withdraws, nil
	}

	withdraws.ResponseStatus = constants.AnswerSuccessfully
	return withdraws, nil
}

func (dbc *DBConnector) VerificationOrderExists(number int) (constants.Answer, error) {

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return constants.AnswerErrorServer, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectAccrualPLUSS, number)
	if err != nil {
		return constants.AnswerErrorServer, err
	}
	defer rows.Close()

	if rows.Next() {
		return constants.AnswerConflict, nil
	}

	return constants.AnswerSuccessfully, nil
}

func (dbc *DBConnector) SetValueScoringSystem(fullScoringSystem *FullScoringSystem) (constants.Answer, error) {

	ctx := context.Background()
	tx, err := dbc.Pool.Begin(ctx)
	if err != nil {
		return constants.AnswerErrorServer, err
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return constants.AnswerErrorServer, err
	}
	defer conn.Release()

	if _, err = conn.Query(ctx, constants.QueryAddAccrual,
		fullScoringSystem.ScoringSystem.Accrual, time.Now(), "PLUS", fullScoringSystem.ScoringSystem.Order); err != nil {

		_ = tx.Rollback(ctx)
		return constants.AnswerErrorServer, err
	}
	conn.Release()

	nameColum := ""
	switch fullScoringSystem.ScoringSystem.Status {
	case "REGISTERED":
		nameColum = "createdAt"
	case "INVALID":
		nameColum = "failedAt"
	case "PROCESSING":
		nameColum = "startedAt"
	case "PROCESSED":
		nameColum = "finishedAt"
	default:
		return constants.AnswerErrorServer, err
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return constants.AnswerErrorServer, err
	}
	defer conn.Release()

	if _, err = conn.Query(ctx,
		fmt.Sprintf(`UPDATE gofermart.orders
					SET "%s"=$2
					WHERE "orderID"=$1;`, nameColum), fullScoringSystem.ScoringSystem.Order, time.Now()); err != nil {

		_ = tx.Rollback(ctx)
		return constants.AnswerErrorServer, err
	}

	_ = tx.Commit(ctx)

	return constants.AnswerSuccessfully, nil
}

func CreateModeLDB(Pool *pgxpool.Pool) {
	ctx := context.Background()
	conn, err := Pool.Acquire(ctx)
	if err != nil {
		return
	}

	if _, err = Pool.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS gofermart`); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS gofermart.order_accrual
								(
									"Accrual" double precision,
									"DateAccrual" timestamp with time zone,
									"TypeAccrual" character varying(10) COLLATE pg_catalog."default",
									"Order" numeric
								)
								
								TABLESPACE pg_default;
								
								ALTER TABLE IF EXISTS gofermart.order_accrual
									OWNER to postgres;
`)
	if err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return
	}

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS gofermart.orders
								(
									"userID" character varying(150) COLLATE pg_catalog."default" NOT NULL,
									"orderID" numeric,
									"createdAt" timestamp with time zone,
									"startedAt" timestamp with time zone,
									"finishedAt" timestamp with time zone,
									"failedAt" timestamp with time zone
								)
								
								TABLESPACE pg_default;
								
								ALTER TABLE IF EXISTS gofermart.orders
									OWNER to postgres;`)
	if err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return
	}

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS gofermart.users
									(
										"User" character varying(150) COLLATE pg_catalog."default" NOT NULL,
										"Password" character varying(256) COLLATE pg_catalog."default" NOT NULL,
										CONSTRAINT users_pkey PRIMARY KEY ("User")
									)
									
									TABLESPACE pg_default;
									
									ALTER TABLE IF EXISTS gofermart.users
										OWNER to postgres;`)
	if err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return
	}
}

func GetOrder4AS(addressAcSys string, number string) (*ScoringSystem, error) {
	addressPost := fmt.Sprintf("%s/api/orders/%s", addressAcSys, number)
	req, err := http.NewRequest("GET", addressPost, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, errors.New("429")
	}

	varsAnswer := mux.Vars(req)
	fmt.Println(varsAnswer)

	body := resp.Body
	contentEncoding := resp.Header.Get("Content-Encoding")
	err = compression.DecompressBody(contentEncoding, body)
	if err != nil {
		return nil, err
	}

	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			return nil, err
		}
		fmt.Println(arrBody)
	}

	scoringSystem := NewScoringService()
	if err = json.NewDecoder(body).Decode(scoringSystem); err != nil {
		return nil, err
	}
	return scoringSystem, nil
}
