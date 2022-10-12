package handlers

import (
	"bytes"
	"encoding/json"
	"gofermart/internal/token"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/theplant/luhn"

	"gofermart/internal/compression"
	"gofermart/internal/constants"
	"gofermart/internal/postgresql"
)

//POST
func (srv *Server) apiUserRegisterPOST(w http.ResponseWriter, r *http.Request) {
	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	respByte, err := ioutil.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
	}

	newAccount := new(postgresql.Account)
	newAccount.Pool = srv.Pool
	if err := json.Unmarshal(respByte, &newAccount.User); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
	}
	newAccount.Key = srv.Cfg.Key

	tx, err := srv.Pool.Begin(srv.Context.Ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rez := newAccount.NewAccount()
	w.WriteHeader(rez)
	if rez == http.StatusOK {
		if err := tx.Commit(srv.Context.Ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}

		tokenString := ""
		ct := token.Claims{Authorized: true, User: newAccount.Name, Exp: constants.TimeLiveToken}
		if tokenString, err = ct.GenerateJWT(); err != nil {
			constants.Logger.ErrorLog(err)
		}
		_, err = w.Write([]byte(tokenString))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}
	} else {
		if err := tx.Rollback(srv.Context.Ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}
	}
}

func (srv *Server) apiUserLoginPOST(w http.ResponseWriter, r *http.Request) {

	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	respByte, err := ioutil.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка чтения тела", http.StatusInternalServerError)
	}

	Account := new(postgresql.Account)
	Account.Pool = srv.Pool
	if err := json.Unmarshal(respByte, &Account.User); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка Unmarshal", http.StatusInternalServerError)
	}
	Account.Key = srv.Cfg.Key

	rez := Account.GetAccount()
	w.WriteHeader(rez)

	if rez != http.StatusOK {
		return
	}

	tokenString := ""
	tc := token.Claims{Authorized: true, User: Account.Name, Exp: constants.TimeLiveToken}
	if tokenString, err = tc.GenerateJWT(); err != nil {
		constants.Logger.ErrorLog(err)
	}
	w.Header().Add("Authorization", tokenString)
	r.Header.Add("Authorization", tokenString)

	_, err = w.Write([]byte(tokenString))
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) apiUserOrdersPOST(w http.ResponseWriter, r *http.Request) {
	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	respByte, err := ioutil.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Error get value", http.StatusInternalServerError)
	}

	numOrder, err := strconv.Atoi(string(respByte))
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "not Luna", http.StatusInternalServerError)
	}

	if !luhn.Valid(numOrder) {
		constants.Logger.ErrorLog(err)
		http.Error(w, "not Luna", http.StatusUnprocessableEntity)
		return
	}

	order := new(postgresql.Order)
	order.Number = numOrder
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	tx, err := srv.Pool.Begin(srv.Context.Ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rez := order.NewOrder()
	w.WriteHeader(rez)
	if rez == http.StatusOK {
		if err := tx.Commit(srv.Context.Ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}
	} else {
		if err := tx.Rollback(srv.Context.Ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}
	}
}

func (srv *Server) apiUserWithdrawPOST(w http.ResponseWriter, r *http.Request) {
	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	respByte, err := ioutil.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка чтения тела", http.StatusInternalServerError)
	}

	orderWithdraw := new(postgresql.OrderWithdraw)
	orderWithdraw.Pool = srv.Pool
	if err := json.Unmarshal(respByte, &orderWithdraw); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка Unmarshal", http.StatusInternalServerError)
		return
	}
	if r.Header["Authorization"] != nil {
		orderWithdraw.Token = r.Header["Authorization"][0]
	}

	w.WriteHeader(orderWithdraw.TryWithdraw())
}

//GET
func (srv *Server) apiUserOrdersGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Number = 0
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	listOrder, status := order.ListOrder()
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, err := w.Write([]byte(""))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}
		return
	}

	listOrderJSON, err := json.MarshalIndent(listOrder, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	w.WriteHeader(status)
	_, err = w.Write(listOrderJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) apiNextStatus(w http.ResponseWriter, r *http.Request) {

	order := new(postgresql.Order)
	order.Number = 0
	order.Pool = srv.Pool

	order.SetNextStatus()

	w.WriteHeader(http.StatusOK)

	if err := r.Body.Close(); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) apiUserBalanceGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Number = 0
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	listBalans, status := order.BalansOrders()
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, err := w.Write([]byte(""))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}

		return
	}

	listBalansJSON, err := json.MarshalIndent(listBalans, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	w.WriteHeader(status)
	_, err = w.Write(listBalansJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) apiUserWithdrawalsGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Number = 0
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	listBalans, status := order.UserWithdrawal()
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, err := w.Write([]byte(""))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}
		return
	}

	listBalansJSON, err := json.MarshalIndent(listBalans, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	w.WriteHeader(status)
	_, err = w.Write(listBalansJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) apiUserAccrualGET(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Number = 0
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	listOrder, status := order.ListOrder()
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, err := w.Write([]byte(""))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}
		return
	}

	listOrderJSON, err := json.MarshalIndent(listOrder, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	//
	w.WriteHeader(status)
	_, err = w.Write(listOrderJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}
