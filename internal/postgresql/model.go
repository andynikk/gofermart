package postgresql

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/cryptography"
	"github.com/andynikk/gofermart/internal/token"
)

type Cfg struct {
	Token string
	*pgxpool.Pool
	Key string `json:"key"`
}

type User struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}

type Account struct {
	User
	*Cfg
}

type Order struct {
	Number string
	*Cfg
}

type OrderWithdraw struct {
	Order    string  `json:"order"`
	Withdraw float64 `json:"sum"`
	*Cfg
}

type orderDB struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC333"`
}

type BalansDB struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
	Total     float64 `json:"total"`
}

type withdrawDB struct {
	Order       int       `json:"order"`
	Withdraw    float64   `json:"sum"`
	DateAccrual time.Time `json:"processed_at" format:"RFC333"`
	Current     float64   `json:"current,omitempty"`
}

type FullScoringSystem struct {
	ScoringSystem *ScoringSystem
	HTTPStatus    int
}

type ScoringSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (a *Account) NewAccount() int {
	ctx := context.Background()
	conn, err := a.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}

	rows, err := conn.Query(ctx, constants.QuerySelectUserWithWhereTemplate, a.Name)
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}
	defer rows.Close()

	if rows.Next() {
		return http.StatusConflict
	}

	heshVal := cryptography.HeshSHA256(a.Password, a.Key)

	if _, err := conn.Exec(ctx, constants.QueryAddUserTemplate, a.Name, heshVal); err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return http.StatusBadRequest
	}
	conn.Release()

	return http.StatusOK
}

func (a *Account) GetAccount() int {
	ctx := context.Background()
	conn, err := a.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer conn.Release()

	heshVal := cryptography.HeshSHA256(a.Password, a.Key)
	rows, err := conn.Query(ctx, constants.QuerySelectUserWithPasswordTemplate, a.Name, heshVal)
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}
	defer rows.Close()

	if rows.Next() {
		return http.StatusOK
	}
	conn.Release()

	return http.StatusUnauthorized
}

func (a *Account) UserOrders() int {
	ctx := context.Background()
	conn, err := a.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryUserOrdersTemplate, a.Name)
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}
	defer rows.Close()

	for rows.Next() {
		return http.StatusConflict
	}

	return http.StatusOK
}

func (o *Order) NewOrder() int {

	ctx := context.Background()
	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return http.StatusUnauthorized
	}

	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, claims["user"], o.Number)
	conn.Release()
	if err != nil {
		return http.StatusBadRequest
	}
	defer rows.Close()

	if rows.Next() {
		return http.StatusOK
	}

	conn, err = o.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}
	if _, err := conn.Exec(ctx, constants.QueryAddOrderTemplate, claims["user"], o.Number, time.Now()); err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return http.StatusBadRequest
	}
	conn.Release()

	return http.StatusAccepted
}

func (o *Order) ListOrder() ([]orderDB, int) {
	var arrOrders []orderDB

	ctx := context.Background()
	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return arrOrders, http.StatusUnauthorized
	}

	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrOrders, http.StatusInternalServerError
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryListOrderTemplate, claims["user"])
	if err != nil {
		return arrOrders, http.StatusBadRequest
	}
	defer rows.Close()

	for rows.Next() {
		var ord orderDB

		err = rows.Scan(&ord.Number, &ord.Status, &ord.Accrual, &ord.UploadedAt)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		arrOrders = append(arrOrders, ord)
	}

	if len(arrOrders) == 0 {
		return arrOrders, http.StatusNoContent
	}

	return arrOrders, http.StatusOK
}

func (o *Order) SetNextStatus() {
	var arrOrders []orderDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return
	}

	rows, err := conn.Query(ctx, constants.QueryListOrderTemplate, "")
	if err != nil {
		conn.Release()
		return
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
			rand.Seed(time.Now().UnixNano())
			randStatus := 2 + rand.Intn(3-2+1)
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
		conn, err := o.Pool.Acquire(ctx)
		if err != nil {
			return
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
			conn, err := o.Pool.Acquire(ctx)
			if err != nil {
				return
			}

			randVal := min + rand.Float64()*(max-min)
			accrual := math.Ceil(randVal*100) / 100
			if _, err = conn.Query(ctx, constants.QueryAddAccrual, &val.Number, accrual, time.Now(), "PLUS"); err != nil {
				constants.Logger.ErrorLog(err)
				continue
			}
			conn.Release()
		}

		conn.Release()
	}
}

type CheckAccrual struct {
	Accrual     float64
	DateAccrual time.Duration
	TypeAccrual string
	Order       int
}

func (o *Order) BalansOrders() (BalansDB, int) {
	var bdb BalansDB

	ctx := context.Background()

	//////////////////////////////////////////////////////////
	//c, e := o.Pool.Acquire(ctx)
	//if e != nil {
	//	return bdb, http.StatusInternalServerError
	//}
	//defer c.Release()
	//r, e := c.Query(ctx, `SELECT "Accrual", "DateAccrual", "TypeAccrual", "Order" FROM gofermart.order_accrual;`)
	//if e != nil {
	//	c.Release()
	//	return bdb, http.StatusBadRequest
	//}
	//defer r.Close()
	//for r.Next() {
	//	var ca CheckAccrual
	//
	//	e = r.Scan(ca.Accrual, ca.DateAccrual, ca.TypeAccrual, ca.Order)
	//	if e != nil {
	//		constants.Logger.ErrorLog(e)
	//	}
	//	fmt.Println(ca)
	//}
	/////////////////////////////////////////////////////

	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return bdb, http.StatusInternalServerError
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return bdb, http.StatusUnauthorized
	}

	rows, err := conn.Query(ctx, constants.QueryUserBalansTemplate, claims["user"])
	if err != nil {
		conn.Release()
		return bdb, http.StatusBadRequest
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return bdb, http.StatusNoContent
		}
	} else {
		return bdb, http.StatusNoContent
	}

	return bdb, http.StatusOK
}

func (o *Order) UserWithdrawal() ([]withdrawDB, int) {
	var arrWithdraw []withdrawDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrWithdraw, http.StatusInternalServerError
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return arrWithdraw, http.StatusUnauthorized
	}

	rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], "MINUS")
	if err != nil {
		conn.Release()
		return arrWithdraw, http.StatusBadRequest
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
		return arrWithdraw, http.StatusNoContent
	}

	return arrWithdraw, http.StatusOK
}

func (o *Order) UserAccrual() ([]withdrawDB, int) {
	var arrWithdraw []withdrawDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrWithdraw, http.StatusInternalServerError
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return arrWithdraw, http.StatusUnauthorized
	}

	//rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], 0, "MINUS")
	rows, err := conn.Query(ctx, constants.QuerySelectAccrual, claims["user"], "MINUS")
	if err != nil {
		conn.Release()
		return arrWithdraw, http.StatusBadRequest
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
		return arrWithdraw, http.StatusNoContent
	}

	return arrWithdraw, http.StatusOK
}

func (ow *OrderWithdraw) TryWithdraw() int {

	ctx := context.Background()
	conn, err := ow.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer conn.Release()

	claims, ok := token.ExtractClaims(ow.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return http.StatusUnauthorized
	}

	rows, err := conn.Query(ctx, constants.QueryOrderBalansTemplate, claims["user"], ow.Order)
	if err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return http.StatusUnprocessableEntity
	}

	sumWithdraw := ow.Withdraw
	for rows.Next() {
		var bdb BalansDB

		err = rows.Scan(&bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		if bdb.Total < sumWithdraw {
			constants.Logger.ErrorLog(err)
			conn.Release()
			return http.StatusPaymentRequired
		}
	}
	conn.Release()

	conn, err = ow.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer conn.Release()
	if _, err = conn.Query(ctx, constants.QueryAddAccrual, ow.Order, sumWithdraw, time.Now(), "MINUS"); err != nil {
		constants.Logger.ErrorLog(err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (cfg *Cfg) ListNotAccrualOrders() ([]orderDB, int) {
	var arrOrderDB []orderDB

	ctx := context.Background()
	conn, err := cfg.Pool.Acquire(ctx)
	if err != nil {
		return arrOrderDB, http.StatusInternalServerError
	}

	rows, err := conn.Query(ctx, `SELECT * FROM gofermart.orders AS orders WHERE `+
		`orders."finishedAt" ISNULL AND orders."failedAt" ISNULL`)
	if err != nil {
		return arrOrderDB, http.StatusInternalServerError
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

	return arrOrderDB, http.StatusOK
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
									"Accrual" numeric,
									"DateAccrual" double precision,
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
									"userID" character(150) COLLATE pg_catalog."default" NOT NULL,
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
