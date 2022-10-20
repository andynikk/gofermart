package handlers

import (
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
			if fullScoringSystem.HttpStatus != http.StatusTooManyRequests {
				data <- fullScoringSystem
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (srv *Server) GetScoringSystem(number string) (fullScoringSystem *postgresql.FullScoringSystem) {
	fullScoringSystem = new(postgresql.FullScoringSystem)
	scoringSystem := new(postgresql.ScoringSystem)

	//fullScoringSystem.HttpStatus = http.StatusTooManyRequests
	//return fullScoringSystem

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

			fullScoringSystem.HttpStatus = http.StatusInternalServerError
			return fullScoringSystem
		}

		arrBody, err := compression.Decompress(bytBody)
		fmt.Println(arrBody)
	}

	if err := json.NewDecoder(bodyJSON).Decode(&scoringSystem); err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		return fullScoringSystem
	}

	fullScoringSystem.ScoringSystem = scoringSystem
	fullScoringSystem.HttpStatus = http.StatusOK
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
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		return
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QuerySelectAccrualPLUSS, order)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	if rows.Next() {
		fullScoringSystem.HttpStatus = http.StatusConflict
		return
	}
	conn.Release()

	tx, err := srv.Pool.Begin(ctx)

	conn, err = srv.Pool.Acquire(ctx)
	if err != nil {
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		_ = tx.Rollback(ctx)
		return
	}
	defer conn.Release()

	if _, err = conn.Query(ctx, constants.QueryAddAccrual,
		order, fullScoringSystem.ScoringSystem.Accrual, time.Now(), "PLUS"); err != nil {

		_ = tx.Rollback(ctx)
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
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
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		return
	}

	conn, err = srv.Pool.Acquire(ctx)
	if err != nil {
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		return
	}
	defer conn.Release()
	rows, err = conn.Query(ctx,
		fmt.Sprintf(`SELECT * FROM gofermart.orders AS orders
							WHERE "orderID"=$1 and "%s" ISNULL;`, nameColum), order)
	defer rows.Close()

	if err != nil {
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		_ = tx.Rollback(ctx)
		constants.Logger.ErrorLog(err)
		return
	}
	if rows.Next() {
		_ = tx.Rollback(ctx)
		return
	}

	if _, err = conn.Query(ctx,
		fmt.Sprintf(`UPDATE gofermart.orders
					SET "%s"=$2
					WHERE "orderID"=$1;`, nameColum), order, time.Now()); err != nil {
		fullScoringSystem.HttpStatus = http.StatusInternalServerError
		constants.Logger.ErrorLog(err)
		_ = tx.Rollback(ctx)
		return
	}
	conn.Release()
	_ = tx.Commit(ctx)

	fullScoringSystem.HttpStatus = http.StatusOK
}
