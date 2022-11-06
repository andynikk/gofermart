package postgresql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/channel"
	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/cryptography"
	"github.com/andynikk/gofermart/internal/token"
	"github.com/andynikk/gofermart/internal/utils"
)

func (dbc *DBConnector) NewAccount(name string, password string) (*User, error) {
	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectUserWithWhereTemplate, name)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer rows.Close()

	user := User{}
	if rows.Next() {
		err := rows.Scan(&user.Name, &user.Password)
		if err != nil {
			return nil, constants.ErrErrorServer
		}
		return nil, constants.ErrLoginBusy
	}

	heshVal := cryptography.HeshSHA256(password, dbc.Cfg.Key)
	if _, err := conn.Exec(ctx, constants.QueryAddUserTemplate, name, heshVal); err != nil {
		return nil, constants.ErrErrorServer
	}
	conn.Release()

	user.Name = name
	user.Password = heshVal

	return &user, nil
}

func (dbc *DBConnector) GetAccount(name string, password string) (*User, error) {

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	heshVal := cryptography.HeshSHA256(password, dbc.Cfg.Key)
	rows, err := conn.Query(ctx, constants.QuerySelectUserWithPasswordTemplate, name, heshVal)
	if err != nil {
		return nil, constants.ErrInvalidFormat
	}
	defer rows.Close()

	user := User{}
	if rows.Next() {
		_ = rows.Scan(&user.Name, &user.Password)
		return &user, nil

	}
	conn.Release()

	return nil, constants.ErrInvalidLoginPassword
}

func (dbc *DBConnector) NewOrder(tkn string, number int) (*Order, error) {
	ctx := context.Background()
	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		return nil, constants.ErrUserNotAuthenticated
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, "", number)
	if err != nil {
		return nil, constants.ErrInvalidFormat
	}
	defer rows.Close()

	order := Order{}
	if rows.Next() {
		_ = rows.Scan(&order.User, &order.Number, &order.CreatedAt,
			&order.StartedAt, &order.FinishedAt, &order.FailedAt, &order.Status)
		if order.User != claims["user"] {
			return nil, constants.ErrUploadedAnotherUser
		}

		return nil, constants.ErrOrderUpload
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, constants.QueryAddOrderTemplate, claims["user"], number, time.Now()); err != nil {
		return nil, constants.ErrInvalidFormat
	}
	/////////////////////////////////////////////////////////////////

	return &order, nil
}

func (dbc *DBConnector) SetStartedAt(number int, tkn string) (*Order, error) {
	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	timeNow := time.Now()
	if _, err = conn.Query(ctx, constants.QueryUpdateStartedAt, timeNow, number); err != nil {
		return nil, constants.ErrInvalidFormat
	}

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		return nil, constants.ErrUserNotAuthenticated
	}

	order := Order{
		Number:    strconv.Itoa(number),
		User:      claims["user"].(string),
		StartedAt: timeNow,
	}

	return &order, nil
}

func (dbc *DBConnector) TryWithdraw(tkn string, number string, sumWithdraw float64) (*Balances, error) {

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		return nil, constants.ErrUserNotAuthenticated
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderBalansTemplate, claims["user"], number)
	if err != nil {
		return nil, constants.ErrErrorServer
	}

	balances := Balances{}
	if rows.Next() {

		err = rows.Scan(&balances.Number, &balances.Total, &balances.Withdrawn, &balances.Current)
		if err != nil {
			return nil, constants.ErrErrorServer
		}
		if balances.Total < sumWithdraw {
			return nil, constants.ErrInsufficientFunds
		}
	}
	conn.Release()

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()
	//TODO: Добавляем спсанные баллы
	if _, err = conn.Query(ctx, constants.QueryAddAccrual, sumWithdraw, time.Now(), "MINUS", number); err != nil {
		return nil, constants.ErrErrorServer
	}
	rows.Close()

	conn.Release()
	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err = conn.Query(ctx, constants.QueryOrderWhereNumTemplate, claims["user"], number)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	if !rows.Next() {
		rows.Close()
		_, err = conn.Query(ctx, constants.QueryAddOrderTemplate, claims["user"], number, time.Now())
		if err != nil {
			return nil, constants.ErrErrorServer
		}
	}

	return &balances, nil
}

func (dbc *DBConnector) ListOrder(tkn string, addressAcSys string) (*OrdersAccrual, error) {
	ctx := context.Background()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		return nil, constants.ErrUserNotAuthenticated
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryListOrderTemplate, claims["user"])
	if err != nil {
		return nil, constants.ErrInvalidFormat
	}
	defer rows.Close()

	var orders OrdersAccrual

	for rows.Next() {
		var order OrderAccrual

		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		fmt.Println()
		ss, err := GetOrder4AS(addressAcSys, order.Number)
		if err == nil {
			order.Status = ss.Status
			order.Accrual = ss.Accrual
		}
		orders.OrderAccrual = append(orders.OrderAccrual, order)
	}
	if len(orders.OrderAccrual) == 0 {
		return nil, constants.ErrNoContent
	}

	return &orders, nil
}

func (dbc *DBConnector) BalancesOrders(tkn string, addressAcSys string) (*Balances, error) {
	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()
	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		return nil, constants.ErrUserNotAuthenticated
	}

	rows, err := conn.Query(ctx, constants.QueryUserOrdes, claims["user"])
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer rows.Close()

	var balances []Balances
	for rows.Next() {
		var bdb Balances

		err = rows.Scan(&bdb.Number, &bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		ss, err := GetOrder4AS(addressAcSys, bdb.Number)
		if err == nil {
			bdb.Current = ss.Accrual
		}
		balances = append(balances, bdb)
	}
	if len(balances) == 0 {
		return nil, constants.ErrNoContent
	}

	var totalBalance Balances
	for _, val := range balances {
		totalBalance.Withdrawn = totalBalance.Withdrawn + val.Withdrawn
		totalBalance.Current = totalBalance.Current + val.Current
	}
	totalBalance.Current = totalBalance.Current - totalBalance.Withdrawn

	if err != nil {
		return nil, constants.ErrErrorServer
	}

	return &totalBalance, nil
}

func (dbc *DBConnector) UserWithdrawal(tkn string) (*Withdraws, error) {

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(tkn)
	if !ok {
		return nil, constants.ErrUserNotAuthenticated
	}

	rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], "MINUS")
	if err != nil {
		//return nil, fmt.Errorf("%w", constants.ErrInvalidFormat)
		return nil, constants.ErrInvalidFormat
	}
	defer rows.Close()

	var withdraws Withdraws
	for rows.Next() {
		var withdraw Withdraw

		err = rows.Scan(&withdraw.Order, &withdraw.DateAccrual, &withdraw.Withdraw, &withdraw.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		withdraws.Withdraw = append(withdraws.Withdraw, withdraw)
	}

	if len(withdraws.Withdraw) == 0 {
		return nil, constants.ErrNoContent
	}

	return &withdraws, nil
}

func (dbc *DBConnector) VerificationOrderExists(number int) error {

	ctx := context.Background()
	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		return constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectAccrualPLUSS, number)
	if err != nil {
		return constants.ErrErrorServer
	}
	defer rows.Close()

	if rows.Next() {
		return constants.ErrConflict
	}

	return nil
}

func (dbc *DBConnector) SetValueScoringOrder(scoringOrder *channel.ScoringOrder) error {

	ctx := context.Background()
	tx, err := dbc.Pool.Begin(ctx)
	if err != nil {
		return constants.ErrErrorServer
	}

	conn, err := dbc.Pool.Acquire(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return constants.ErrErrorServer
	}
	defer conn.Release()

	if _, err = conn.Query(ctx, constants.QueryAddAccrual,
		scoringOrder.Accrual, time.Now(), "PLUS", scoringOrder.Order); err != nil {

		_ = tx.Rollback(ctx)
		return constants.ErrErrorServer
	}
	conn.Release()

	nameColum := ""
	switch scoringOrder.Status {
	case "REGISTERED":
		nameColum = "createdAt"
	case "INVALID":
		nameColum = "failedAt"
	case "PROCESSING":
		nameColum = "startedAt"
	case "PROCESSED":
		nameColum = "finishedAt"
	default:
		return constants.ErrErrorServer
	}

	conn, err = dbc.Pool.Acquire(ctx)
	if err != nil {
		return constants.ErrErrorServer
	}
	defer conn.Release()

	if _, err = conn.Query(ctx,
		fmt.Sprintf(`UPDATE gofermart.orders
					SET "%s"=$2
					WHERE "orderID"=$1;`, nameColum), scoringOrder.Order, time.Now()); err != nil {

		_ = tx.Rollback(ctx)
		return constants.ErrInvalidFormat
	}

	_ = tx.Commit(ctx)

	return nil
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

func GetOrder4AS(addressAcSys string, number string) (*channel.ScoringOrder, error) {
	addressPost := fmt.Sprintf("%s/api/orders/%s", addressAcSys, number)
	resp, err := utils.GETQuery(addressPost)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		return nil, errors.New("429")
	}

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

	ScoringOrder := channel.NewScoringOrder()
	if err = json.NewDecoder(body).Decode(ScoringOrder); err != nil {
		return nil, err
	}
	return ScoringOrder, nil
}
