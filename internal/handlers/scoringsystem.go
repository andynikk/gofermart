package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/postgresql"
)

func (srv *Server) ScoringSystem(number string, data chan *postgresql.FullScoringSystem) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	//getTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-data:
			cancelFunc()
			return
		case <-ctx.Done():
			cancelFunc()
			return
		default:
			fullScoringSystem := srv.GetScoringSystem(number)
			if fullScoringSystem.HTTPStatus != http.StatusTooManyRequests {
				data <- fullScoringSystem
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (srv *Server) GetScoringSystem(number string) (fullScoringSystem *postgresql.FullScoringSystem) {
	fullScoringSystem = new(postgresql.FullScoringSystem)
	scoringSystem := new(postgresql.ScoringSystem)

	addressPost := fmt.Sprintf("http://%s/api/orders/%s", srv.AddressAcSys, number)
	req, err := http.NewRequest("GET", addressPost, strings.NewReader(""))
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer resp.Body.Close()

	varsAnswer := mux.Vars(req)
	fmt.Println(varsAnswer)

	bodyJSON := resp.Body

	contentEncoding := resp.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(resp.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)

			fullScoringSystem.HTTPStatus = http.StatusInternalServerError
			return fullScoringSystem
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)

			fullScoringSystem.HTTPStatus = http.StatusInternalServerError
			return fullScoringSystem
		}
		fmt.Println(arrBody)
	}

	if err := json.NewDecoder(bodyJSON).Decode(&scoringSystem); err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		return fullScoringSystem
	}

	fullScoringSystem.ScoringSystem = scoringSystem
	fullScoringSystem.HTTPStatus = http.StatusOK
	return fullScoringSystem
}

func (srv *Server) SetValueScoringSystem(fullScoringSystem *postgresql.FullScoringSystem) {

	order, err := strconv.Atoi(fullScoringSystem.ScoringSystem.Order)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	ctx := context.Background()
	conn, err := srv.Pool.Acquire(ctx)
	fmt.Println("-----3")
	if err != nil {
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		return
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectAccrualPLUSS, order)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	if rows.Next() {
		fullScoringSystem.HTTPStatus = http.StatusConflict
		return
	}
	conn.Release()

	tx, err := srv.Pool.Begin(ctx)
	if err != nil {
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		return
	}

	conn, err = srv.Pool.Acquire(ctx)
	if err != nil {
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		_ = tx.Rollback(ctx)
		return
	}
	defer conn.Release()

	if _, err = conn.Query(ctx, constants.QueryAddAccrual,
		order, fullScoringSystem.ScoringSystem.Accrual, time.Now(), "PLUS"); err != nil {

		_ = tx.Rollback(ctx)
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		constants.Logger.ErrorLog(err)
		return
	}

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
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		return
	}

	conn, err = srv.Pool.Acquire(ctx)
	if err != nil {
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		return
	}
	defer conn.Release()
	rows, err = conn.Query(ctx,
		fmt.Sprintf(`SELECT * FROM gofermart.orders AS orders
							WHERE "orderID"=$1 and "%s" ISNULL;`, nameColum), order)
	if err != nil {
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		_ = tx.Rollback(ctx)
		constants.Logger.ErrorLog(err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		_ = tx.Rollback(ctx)
		return
	}

	if _, err = conn.Query(ctx,
		fmt.Sprintf(`UPDATE gofermart.orders
					SET "%s"=$2
					WHERE "orderID"=$1;`, nameColum), order, time.Now()); err != nil {
		fullScoringSystem.HTTPStatus = http.StatusInternalServerError
		constants.Logger.ErrorLog(err)
		_ = tx.Rollback(ctx)
		return
	}
	conn.Release()
	_ = tx.Commit(ctx)

	fullScoringSystem.HTTPStatus = http.StatusOK
}

type Goods struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}

func (srv *Server) AddItemsScoringSystem(good *Goods) {

	jsonStr, err := json.MarshalIndent(good, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	addressPost := fmt.Sprintf("http://%s/api/goods", srv.AddressAcSys)
	req, err := http.NewRequest("POST", addressPost, bytes.NewBuffer(jsonStr))
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}

type GoodOrderSS struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type OrderSS struct {
	Order       string        `json:"order"`
	GoodOrderSS []GoodOrderSS `json:"goods"`
}

func (srv *Server) AddOrderScoringSystem(orderSS *OrderSS) {

	jsonStr, err := json.MarshalIndent(orderSS, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	bufJsonStr := bytes.NewBuffer(jsonStr)
	addressPost := fmt.Sprintf("http://%s/api/orders", srv.AddressAcSys)
	req, err := http.NewRequest("POST", addressPost, bufJsonStr)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer resp.Body.Close()

	//resp, err := http.Post(addressPost, "application/json", bufJsonStr)
	//if err != nil {
	//	constants.Logger.ErrorLog(err)
	//	return
	//}
	fmt.Println("response Status:", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}
