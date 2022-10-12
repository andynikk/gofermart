package postgresql

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"gofermart/internal/token"
	"math"
	"math/rand"
	"net/http"
	"time"

	"gofermart/internal/constants"
	"gofermart/internal/cryptography"
)

type User struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}

type Cfg struct {
	Key string `json:"key"`
}

type Account struct {
	User
	Cfg
	*pgxpool.Pool
}

type Order struct {
	Number int
	Token  string
	*pgxpool.Pool
}

type OrderWithdraw struct {
	Order    int     `json:"order"`
	Withdraw float64 `json:"sum"`
	Token    string
	*pgxpool.Pool
}

type orderDB struct {
	Order      int       `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC333"`
}

type balansDB struct {
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

func (a *Account) NewAccount() int {
	ctx := context.Background()
	conn, err := a.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}

	rows, err := conn.Query(ctx, constants.QuerySelectUserWithWhereTemplate, a.Name)
	defer rows.Close()
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}

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
	defer rows.Close()
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}

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

	rows, err := conn.Query(ctx, constants.QueryUserOrdersTemplate, a.Name)
	defer rows.Close()
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}

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
	rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, o.Number)
	defer rows.Close()
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}

	if rows.Next() {
		return http.StatusConflict
	}
	if _, err := conn.Exec(ctx, constants.QueryAddOrderTemplate, claims["user"], o.Number); err != nil {
		constants.Logger.ErrorLog(err)
		conn.Release()
		return http.StatusBadRequest
	}

	rows, err = conn.Query(ctx, constants.QueryAddStatusTemplate, o.Number, "NEW", time.Now())
	defer rows.Close()
	if err != nil {
		conn.Release()
		return http.StatusBadRequest
	}
	conn.Release()

	return http.StatusOK
}

func (o *Order) ListOrder() ([]orderDB, int) {
	var arrOrders []orderDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrOrders, http.StatusInternalServerError
	}

	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return arrOrders, http.StatusUnauthorized
	}

	rows, err := conn.Query(ctx, constants.QueryListOrderWhereTemplate, claims["user"])
	defer rows.Close()
	if err != nil {
		conn.Release()
		return arrOrders, http.StatusBadRequest
	}
	for rows.Next() {
		var ord orderDB

		err = rows.Scan(&ord.Order, &ord.Status, &ord.Accrual, &ord.UploadedAt)
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

	rows, err := conn.Query(ctx, constants.QueryListOrderTemplate)
	defer rows.Close()
	if err != nil {
		conn.Release()
		return
	}
	for rows.Next() {
		var ord orderDB

		err = rows.Scan(&ord.Order, &ord.Status, &ord.Accrual, &ord.UploadedAt)
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
				nextStatus = constants.StatusINVALID.String()
			} else {
				nextStatus = constants.StatusPROCESSED.String()
			}
		default:
			nextStatus = constants.StatusPROCESSING.String()
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
			`INSERT INTO gofermart.order_statuses("Order", "Status", "DateStatus")
					VALUES ($1, $2, $3);`, &val.Order, &val.Status, time.Now()); err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		conn.Release()

		if val.Status == "PROCESSED" {
			conn, err := o.Pool.Acquire(ctx)
			if err != nil {
				return
			}

			randVal := min + rand.Float64()*(max-min)
			accrual := math.Ceil(randVal*100) / 100
			if _, err = conn.Query(ctx, constants.QueryAddAccrual, &val.Order, accrual, time.Now(), "PLUS"); err != nil {
				constants.Logger.ErrorLog(err)
				continue
			}
			conn.Release()
		}

		conn.Release()
	}
}

func (o *Order) BalansOrders() ([]balansDB, int) {
	var arrBalans []balansDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrBalans, http.StatusInternalServerError
	}

	claims, ok := token.ExtractClaims(o.Token)
	if !ok {
		constants.Logger.InfoLog("error extract claims")
		return arrBalans, http.StatusUnauthorized
	}

	rows, err := conn.Query(ctx, constants.QueryUserBalansTemplate, claims["user"])
	defer rows.Close()
	if err != nil {
		conn.Release()
		return arrBalans, http.StatusBadRequest
	}
	for rows.Next() {
		var bdb balansDB

		err = rows.Scan(&bdb.Total, &bdb.Withdrawn, &bdb.Current)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		arrBalans = append(arrBalans, bdb)
	}

	if len(arrBalans) == 0 {
		return arrBalans, http.StatusNoContent
	}

	return arrBalans, http.StatusOK
}

func (o *Order) UserWithdrawal() ([]withdrawDB, int) {
	var arrWithdraw []withdrawDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrWithdraw, http.StatusInternalServerError
	}

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

func (o *Order) UserAccrual() ([]withdrawDB, int) {
	var arrWithdraw []withdrawDB

	ctx := context.Background()
	conn, err := o.Pool.Acquire(ctx)
	if err != nil {
		return arrWithdraw, http.StatusInternalServerError
	}

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
	var arrBalans []balansDB

	ctx := context.Background()
	conn, err := ow.Pool.Acquire(ctx)
	if err != nil {
		return http.StatusInternalServerError
	}

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
		var bdb balansDB

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
		arrBalans = append(arrBalans, bdb)
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
