package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/andynikk/gofermart/internal/channel"
	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/random"
	"github.com/andynikk/gofermart/internal/utils"
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

func (srv *Server) ScoringOrder(number string) {

	ctx, cancelFunc := context.WithCancel(context.Background())

	for {
		select {
		case <-srv.ChanData:
			cancelFunc()
			return
		case <-ctx.Done():
			cancelFunc()
			return
		default:
			fss, err := srv.GetScoringOrder(number)
			if !errors.Is(err, constants.ErrTooManyRequests) {
				srv.ChanData <- fss
			} else {
				time.Sleep(1 * time.Second)
			}

		}
	}
}

func (srv *Server) GetScoringOrder(number string) (*channel.ScoringOrder, error) {
	ctx := context.Background()
	conn, err := srv.Pool.Acquire(ctx)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, constants.QueryOrderWhereNumTemplate, "", number)
	conn.Release()
	if err != nil {
		return nil, constants.ErrInvalidFormat
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, constants.ErrInvalidOrderNumber
	}

	addressPost := fmt.Sprintf("%s/api/orders/%s", srv.AccrualAddress, number)
	resp, err := utils.GETQuery(addressPost)
	if err != nil {
		return nil, constants.ErrErrorServer
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, constants.ErrTooManyRequests
	}

	body := resp.Body
	contentEncoding := resp.Header.Get("Content-Encoding")
	err = compression.DecompressBody(contentEncoding, body)
	if err != nil {
		return nil, constants.ErrErrorServer
	}

	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, constants.ErrErrorServer
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			return nil, constants.ErrErrorServer
		}
		fmt.Println(arrBody)
	}

	scoringOrder := channel.NewScoringOrder()
	if err = json.NewDecoder(body).Decode(scoringOrder); err != nil {
		return nil, constants.ErrErrorServer
	}

	return scoringOrder, nil
}

func (srv *Server) SetValueScoringOrder(scoringOrder *channel.ScoringOrder) error {
	order, err := strconv.Atoi(scoringOrder.Order)
	if err != nil {
		return err
	}

	if err := srv.DBConnector.VerificationOrderExists(order); err != nil {
		return err
	}

	err = srv.DBConnector.SetValueScoringOrder(scoringOrder)
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) AddItemsScoringOrder(good *Goods) {

	jsonStr, err := json.MarshalIndent(good, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	addressPost := fmt.Sprintf("%s/api/goods", srv.AccrualAddress)
	resp, err := utils.POSTQuery(addressPost, jsonStr)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}

func (srv *Server) AddOrderScoringOrder(orderSS *OrderSS) error {

	jsonStr, err := json.MarshalIndent(orderSS, "", " ")
	if err != nil {
		return err
	}

	addressPost := fmt.Sprintf("%s/api/orders", srv.AccrualAddress)
	resp, err := utils.POSTQuery(addressPost, jsonStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func NewGoodOrderSS() *GoodOrderSS {
	return &GoodOrderSS{
		Description: random.RandNameItem(2, 3),
		Price:       random.RandPriceItem(1000.00, 3000.00),
	}
}
