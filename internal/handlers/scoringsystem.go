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

type Goods struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}

type GoodOrderSS struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type OrderSS struct {
	Order       string        `json:"order"`
	GoodOrderSS []GoodOrderSS `json:"goods"`
}

func (srv *Server) ScoringSystem(number string, data chan *postgresql.FullScoringSystem) {

	ctx, cancelFunc := context.WithCancel(context.Background())

	for {
		select {
		case <-data:
			cancelFunc()
			return
		case <-ctx.Done():
			cancelFunc()
			return
		default:
			fss, _ := srv.GetScoringSystem(number)
			if fss.Answer != constants.AnswerTooManyRequests {
				data <- fss
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (srv *Server) GetScoringSystem(number string) (*postgresql.FullScoringSystem, error) {
	fullScoringSystem := new(postgresql.FullScoringSystem)
	ScoringSystem := new(postgresql.ScoringSystem)

	fullScoringSystem.ScoringSystem = ScoringSystem

	ctx := context.Background()
	conn, err := srv.Pool.Acquire(ctx)
	if err != nil {
		return fullScoringSystem, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, "", number)
	conn.Release()
	if err != nil {
		return fullScoringSystem, err
	}
	defer rows.Close()

	if !rows.Next() {
		fullScoringSystem.Answer = constants.AnswerInvalidOrderNumber
		return fullScoringSystem, nil
	}

	addressPost := fmt.Sprintf("http://%s/api/orders/%s", srv.AddressAcSys, number)
	req, err := http.NewRequest("GET", addressPost, strings.NewReader(""))
	if err != nil {
		return fullScoringSystem, err
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fullScoringSystem, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		fullScoringSystem.Answer = constants.AnswerTooManyRequests
		return fullScoringSystem, nil
	}

	varsAnswer := mux.Vars(req)
	fmt.Println(varsAnswer)

	body := resp.Body
	contentEncoding := resp.Header.Get("Content-Encoding")
	err = compression.DecompressBody(contentEncoding, body)
	if err != nil {
		return fullScoringSystem, err
	}

	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fullScoringSystem, err
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			return fullScoringSystem, err
		}
		fmt.Println(arrBody)
	}

	if err = json.NewDecoder(body).Decode(fullScoringSystem.ScoringSystem); err != nil {
		return fullScoringSystem, err
	}

	fullScoringSystem.Answer = constants.AnswerSuccessfully
	return fullScoringSystem, nil
}

func (srv *Server) SetValueScoringSystem(fullScoringSystem *postgresql.FullScoringSystem) error {
	order, err := strconv.Atoi(fullScoringSystem.ScoringSystem.Order)
	if err != nil {
		return err
	}

	answer, err := srv.DBConnector.VerificationOrderExists(order)
	if err != nil {
		return err
	}
	if answer.Answer != constants.AnswerSuccessfully {
		return nil
	}

	answer, err = srv.DBConnector.SetValueScoringSystem(fullScoringSystem)
	if err != nil {
		return err
	}

	fullScoringSystem.Answer = answer.Answer
	return nil
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

func (srv *Server) AddOrderScoringSystem(orderSS *OrderSS) error {

	jsonStr, err := json.MarshalIndent(orderSS, "", " ")
	if err != nil {
		return err
	}

	bufJSONStr := bytes.NewBuffer(jsonStr)
	addressPost := fmt.Sprintf("http://%s/api/orders", srv.AddressAcSys)
	req, err := http.NewRequest("POST", addressPost, bufJSONStr)
	if err != nil {
		return err
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	answerBD := new(postgresql.AnswerBD)
	answerBD.Answer = constants.AnswerSuccessfully
	return nil
}
