package postgresql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andynikk/gofermart/internal/compression"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/cryptography"
	"github.com/andynikk/gofermart/internal/random"
	"github.com/andynikk/gofermart/internal/token"
)

func (dbc *DBConnector) NewAccount(name string, password string) (*Account, error) {
	account := new(Account)
	account.User = new(User)

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
	account := new(Account)
	account.User = new(User)

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
	order := new(Order)
	order.OrderUser = new(OrderUser)

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
	order := new(Order)

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	timeNow := time.Now()
	rows, err := conn.Query(ctx, constants.QueryUpdateStartedAt, timeNow, number)
	if err != nil {
		conn.Release()
		return nil, err
	}
	defer rows.Close()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		order.Number = strconv.Itoa(number)
		order.StartedAt = timeNow
		order.ResponseStatus = constants.AnswerUserNotAuthenticated

		return order, nil
	}

	nameUser := claims["user"].(string)
	order.Number = strconv.Itoa(number)
	order.User = nameUser
	order.StartedAt = timeNow
	order.ResponseStatus = constants.AnswerSuccessfully

	return order, nil
}

func (dbc *DBConnector) TryWithdraw(tkn string, number string, sumWithdraw float64) (*AnswerBD, error) {
	answerBD := new(AnswerBD)

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		answerBD.Answer = constants.AnswerUserNotAuthenticated
		return answerBD, nil
	}

	//rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, claims["user"], number)
	//if err != nil {
	//	constants.Logger.ErrorLog(err)
	//	return nil, err
	//}
	//if !rows.Next() {
	//	answerBD.Answer = constants.AnswerInvalidOrderNumber
	//	return answerBD, nil
	//}
	//conn.Release()

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderBalansTemplate, claims["user"], number)
	if err != nil {
		fmt.Println("--------------1", err)
		fmt.Println("--------------1", err.Error())
		return nil, err
	}

	for rows.Next() {
		var bdb BalanceDB

		err = rows.Scan(&bdb.Number, &bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			fmt.Println("--------------2", err)
			fmt.Println("--------------2", err.Error())
			return nil, err
		}
		if bdb.Total < sumWithdraw {
			answerBD.Answer = constants.AnswerInsufficientFunds
			return answerBD, nil
		}
	}
	conn.Release()

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	//TODO: Добавляем спсанные баллы
	fmt.Println("--------------4.3-начало добавления списания")
	fmt.Println("--------------4.3-", sumWithdraw, time.Now(), "MINUS", number)
	if _, err = conn.Query(ctx, constants.QueryAddAccrual, sumWithdraw, time.Now(), "MINUS", number); err != nil {
		fmt.Println("--------------4.3", err)
		fmt.Println("--------------4.3", err.Error())
		return nil, err
	}
	fmt.Println("--------------4.3-окончание добавления списания")
	rows.Close()

	fmt.Println("--------------4.4-проверка на существование ордера")
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
		fmt.Println("--------------4.4-ордер не существует. создаем")
		rows.Close()
		_, err = conn.Query(ctx, constants.QueryAddOrderTemplate, claims["user"], number, time.Now())
		if err != nil {
			return nil, err
		}
	}

	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) ListOrder(tkn string, addressAcSys string) (*AnswerBD, error) {
	answerBD := new(AnswerBD)

	ctx := context.Background()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		answerBD.Answer = constants.AnswerUserNotAuthenticated
		return answerBD, nil
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

	var arrOrders []orderDB
	for rows.Next() {
		var ord orderDB

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
		arrOrders = append(arrOrders, ord)
	}

	if len(arrOrders) == 0 {
		answerBD.Answer = constants.AnswerNoContent
		return answerBD, nil
	}

	listOrderJSON, err := json.MarshalIndent(arrOrders, "", " ")
	if err != nil {
		return nil, err
	}
	answerBD.JSON = listOrderJSON
	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) BalansOrders(tkn string, addressAcSys string) (*AnswerBD, error) {
	answerBD := new(AnswerBD)
	fmt.Println("++++++++++++++++++8.1-")
	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	fmt.Println("++++++++++++++++++8.2-")
	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		answerBD.Answer = constants.AnswerUserNotAuthenticated
		return answerBD, nil
	}

	//rows, err := conn.Query(ctx, constants.QueryUserBalansTemplate, claims["user"])
	rows, err := conn.Query(ctx, constants.QueryUserOrdes, claims["user"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fmt.Println("++++++++++++++++++8.3-")
	var arrBalance []BalanceDB
	for rows.Next() {
		var bdb BalanceDB

		err = rows.Scan(&bdb.Number, &bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			fmt.Println("++++++++++++++++++8.3-err", err)
			constants.Logger.ErrorLog(err)
			continue
		}
		fmt.Println("++++++++++++++++++8.3-строка", bdb)
		ss, err := GetOrder4AS(addressAcSys, bdb.Number)
		if err == nil {
			fmt.Println("++++++++++++++++++8.3-SS", ss.Accrual)
			bdb.Current = ss.Accrual
		}
		arrBalance = append(arrBalance, bdb)
	}
	fmt.Println("++++++++++++++++++8.3+")
	fmt.Println("++++++++++++++++++8.4-")
	if len(arrBalance) == 0 {
		fmt.Println("++++++++++++++++++8-пусто")
		answerBD.Answer = constants.AnswerNoContent
		return answerBD, nil
	}

	var tbdb totalBalanceDB
	for _, val := range arrBalance {
		tbdb.Withdrawn = tbdb.Withdrawn + val.Withdrawn
		tbdb.Current = tbdb.Current + val.Current
	}
	tbdb.Current = tbdb.Current - tbdb.Withdrawn

	fmt.Println("++++++++++++++++++8.5-")
	if err != nil {
		return nil, err
	}

	listBalansJSON, err := json.MarshalIndent(tbdb, "", " ")
	if err != nil {
		return nil, err
	}
	answerBD.JSON = listBalansJSON
	answerBD.Answer = constants.AnswerSuccessfully
	fmt.Println("++++++++++++++++++8.6-")
	return answerBD, nil
}

func (dbc *DBConnector) UserWithdrawal(tkn string) (*AnswerBD, error) {
	answerBD := new(AnswerBD)

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		answerBD.Answer = constants.AnswerUserNotAuthenticated
		return answerBD, nil
	}

	rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], "MINUS")
	if err != nil {
		answerBD.Answer = constants.AnswerInvalidFormat
		return answerBD, nil
	}
	defer rows.Close()

	var arrWithdraw []withdrawDB
	for rows.Next() {
		var bdb withdrawDB

		err = rows.Scan(&bdb.Order, &bdb.DateAccrual, &bdb.Withdraw, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		arrWithdraw = append(arrWithdraw, bdb)
	}

	if len(arrWithdraw) == 0 {
		answerBD.Answer = constants.AnswerNoContent
		return answerBD, nil
	}

	listWithdrawalJSON, err := json.MarshalIndent(arrWithdraw, "", " ")
	if err != nil {
		return nil, err
	}
	answerBD.JSON = listWithdrawalJSON
	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) UserOrders(name string) (*AnswerBD, error) {
	answerBD := new(AnswerBD)

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryUserOrdersTemplate, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		answerBD.Answer = constants.AnswerConflict
		return answerBD, nil
	}

	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) SetNextStatus() (*AnswerBD, error) {
	var arrOrders []orderDB
	fmt.Println("************************1")
	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryListOrderTemplate, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ord orderDB

		err = rows.Scan(&ord.Number, &ord.Status, &ord.Accrual, &ord.UploadedAt)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}

		nextStatus := ""
		switch ord.Status {
		case constants.StatusPROCESSING.String():
			randStatus := random.RandInt(2, 3)
			if randStatus == 3 {
				nextStatus = "failedAt"
			} else {
				nextStatus = "finishedAt"
			}
		default:
			nextStatus = "startedAt"
		}
		ord.Status = nextStatus
		arrOrders = append(arrOrders, ord)
	}
	conn.Release()

	min := 100.10
	max := 501.98

	for _, val := range arrOrders {
		conn, err := dbc.Pool.Acquire(ctx)
		if err != nil {
			return nil, err
		}
		if _, err = conn.Query(ctx,
			fmt.Sprintf(`UPDATE gofermart.orders
					SET "%s"=$2
					WHERE "orderID"=$1;`, val.Status), &val.Number, time.Now()); err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		conn.Release()

		if val.Status == "finishedAt" {
			conn, err := dbc.Pool.Acquire(ctx)
			if err != nil {
				return nil, err
			}
			defer conn.Release()

			accrual := random.RandPriceItem(min, max)
			if _, err = conn.Query(ctx, constants.QueryAddAccrual, &val.Number, accrual, time.Now(), "PLUS"); err != nil {
				constants.Logger.ErrorLog(err)
				continue
			}
			conn.Release()
		}
	}

	answerBD := new(AnswerBD)
	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) UserAccrual(tkn string) (*AnswerBD, error) {
	answerBD := new(AnswerBD)
	var arrWithdraw []withdrawDB

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		answerBD.Answer = constants.AnswerUserNotAuthenticated
		return answerBD, nil
	}

	rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], "MINUS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bdb withdrawDB

		err = rows.Scan(&bdb.Order, &bdb.DateAccrual, &bdb.Withdraw, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		arrWithdraw = append(arrWithdraw, bdb)
	}

	if len(arrWithdraw) == 0 {
		answerBD.Answer = constants.AnswerNoContent
		return answerBD, nil
	}

	listWithdrawalJSON, err := json.MarshalIndent(arrWithdraw, "", " ")
	if err != nil {
		return nil, err
	}
	answerBD.JSON = listWithdrawalJSON
	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) ListNotAccrualOrders() (*AnswerBD, error) {
	var arrOrderDB []orderDB

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, `SELECT * FROM gofermart.orders AS orders WHERE `+
		`orders."finishedAt" ISNULL AND orders."failedAt" ISNULL`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ord orderDB

		err = rows.Scan(&ord.Number, &ord.Status, &ord.Accrual)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		if ord.Status == "REGISTERED" {
			ord.Status = "NEW"
		}
		arrOrderDB = append(arrOrderDB, ord)
	}

	answerBD := new(AnswerBD)
	listWithdrawalJSON, err := json.MarshalIndent(arrOrderDB, "", " ")
	if err != nil {
		return nil, err
	}
	answerBD.JSON = listWithdrawalJSON
	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, nil
}

func (dbc *DBConnector) VerificationOrderExists(number int) (*AnswerBD, error) {
	answerBD := new(AnswerBD)

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectAccrualPLUSS, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		answerBD.Answer = constants.AnswerConflict
		return answerBD, err
	}

	answerBD.Answer = constants.AnswerSuccessfully
	return answerBD, err
}

func (dbc *DBConnector) SetValueScoringSystem(fullScoringSystem *FullScoringSystem) (*AnswerBD, error) {

	fmt.Println("************************2")
	answer := new(AnswerBD)

	ctx := context.Background()
	tx, err := dbc.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	defer conn.Release()

	if _, err = conn.Query(ctx, constants.QueryAddAccrual,
		fullScoringSystem.ScoringSystem.Order, fullScoringSystem.ScoringSystem.Accrual, time.Now(), "PLUS"); err != nil {

		_ = tx.Rollback(ctx)
		return nil, err
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
		err := errors.New("непределен статус")
		return nil, err
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if _, err = conn.Query(ctx,
		fmt.Sprintf(`UPDATE gofermart.orders
					SET "%s"=$2
					WHERE "orderID"=$1;`, nameColum), fullScoringSystem.ScoringSystem.Order, time.Now()); err != nil {

		_ = tx.Rollback(ctx)
		return nil, err
	}

	_ = tx.Commit(ctx)

	answer.Answer = constants.AnswerSuccessfully
	return answer, nil
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

	scoringSystem := new(ScoringSystem)
	if err = json.NewDecoder(body).Decode(scoringSystem); err != nil {
		return nil, err
	}
	return scoringSystem, nil
}
