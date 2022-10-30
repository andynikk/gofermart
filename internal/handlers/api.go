package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/andynikk/gofermart/internal/random"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/theplant/luhn"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/postgresql"
	"github.com/andynikk/gofermart/internal/token"
)

// POST
// 1 TODO: Регистрация пользователя
func (srv *Server) apiUserRegisterPOST(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	contentEncoding := r.Header.Get("Content-Encoding")

	err := compression.DecompressBody(contentEncoding, body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respByte, err := io.ReadAll(body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	Account := postgresql.NewAccount()
	if err := Account.FromJSON(respByte); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 1.2 TODO: Регистрация пользователя в БД.
	// 1.2.1 TODO: Ищем пользовотеля в таблице БД. Если находим, то не создаем. Пароль кэшируется
	account, err := srv.DBConnector.NewAccount(Account.Name, Account.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokenString := ""
	if account.ResponseStatus != constants.AnswerSuccessfully {
		w.Header().Add("Authorization", tokenString)
		w.WriteHeader(HTTPAnswer(account.ResponseStatus))
		return
	}

	// 1.3 TODO: Создание токена
	// 1.3.1 TODO: Если пользователь добавлен создаем токен
	ct := token.Claims{Authorized: true, User: Account.Name, Exp: constants.TimeLiveToken}
	if tokenString, err = ct.GenerateJWT(); err != nil {
		w.Header().Add("Authorization", "")
		http.Error(w, "Ошибка получения токена", http.StatusInternalServerError)
		return
	}

	// 1.4 TODO: Добавление токена в Header
	w.Header().Add("Authorization", tokenString)
	w.WriteHeader(HTTPAnswer(account.ResponseStatus))
}

// 2 TODO: Аутентификации пользователя
func (srv *Server) apiUserLoginPOST(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	contentEncoding := r.Header.Get("Content-Encoding")

	err := compression.DecompressBody(contentEncoding, body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respByte, err := io.ReadAll(body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	Account := postgresql.NewAccount()
	if err := Account.FromJSON(respByte); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 2.1 TODO: Аутентификации пользователя в БД
	account, err := srv.DBConnector.GetAccount(Account.Name, Account.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString := ""
	if account.ResponseStatus != constants.AnswerSuccessfully {
		w.Header().Add("Authorization", tokenString)
		w.WriteHeader(HTTPAnswer(account.ResponseStatus))
		return
	}

	// 2.2 TODO: Создание токена
	tc := token.Claims{Authorized: true, User: Account.Name, Exp: constants.TimeLiveToken}
	if tokenString, err = tc.GenerateJWT(); err != nil {
		w.Header().Add("Authorization", "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2.2 TODO: Добавление токена в Header
	w.Header().Add("Authorization", tokenString)
	w.WriteHeader(HTTPAnswer(account.ResponseStatus))
}

// 3 TODO: Добавление нового ордера
func (srv *Server) apiUserOrdersPOST(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	contentEncoding := r.Header.Get("Content-Encoding")

	err := compression.DecompressBody(contentEncoding, body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respByte, err := io.ReadAll(body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	numOrder, err := strconv.Atoi(string(respByte))
	if err != nil || numOrder == 0 {
		constants.Logger.ErrorLog(err)
		http.Error(w, "bad number order", http.StatusUnprocessableEntity)
		return
	}

	//TODO: Проверка на Луна
	if !luhn.Valid(numOrder) {
		http.Error(w, "bad number order", http.StatusUnprocessableEntity)
		return
	}

	tokenHeader := ""
	if r.Header["Authorization"] != nil {
		tokenHeader = r.Header["Authorization"][0]
	}

	// 3.1 TODO: Добавление нового ордера в БД.
	// 3.1.1 TODO: Ищем ордер по номеру. Если не находим, то создаем
	order, err := srv.DBConnector.NewOrder(tokenHeader, numOrder)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(HTTPAnswer(order.ResponseStatus))
	if order.ResponseStatus != constants.AnswerAccepted {
		w.WriteHeader(HTTPAnswer(order.ResponseStatus))
	}

	if srv.DemoMode {
		min := 1000.00
		max := 3000.00

		goodOrderSS := NewGoodOrderSS()
		goodOrderSS.Description = random.RandNameItem(2, 3)
		goodOrderSS.Price = random.RandPriceItem(min, max)

		var arrGoodOrderSS []GoodOrderSS
		arrGoodOrderSS = append(arrGoodOrderSS, *goodOrderSS)

		orderSS := OrderSS{
			string(respByte),
			arrGoodOrderSS,
		}
		err = srv.AddOrderScoringSystem(&orderSS)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}

		order, err = srv.DBConnector.SetStartedAt(numOrder, tokenHeader)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
		w.WriteHeader(HTTPAnswer(order.ResponseStatus))
	}
}

// 4 TODO: Списание баллов лояльности
func (srv *Server) apiUserWithdrawPOST(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	contentEncoding := r.Header.Get("Content-Encoding")

	err := compression.DecompressBody(contentEncoding, body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
	}

	respByte, err := io.ReadAll(body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка чтения тела", http.StatusInternalServerError)
	}

	orderWithdraw := postgresql.NewOrderWithdraw()
	if err := orderWithdraw.FromJSON(respByte); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка Unmarshal", http.StatusInternalServerError)
		return
	}

	fmt.Println(orderWithdraw)

	tokenHeader := ""
	if r.Header["Authorization"] != nil {
		tokenHeader = r.Header["Authorization"][0]
	}

	// 4.1 TODO: Списание баллов лояльности в БД
	// 4.1.1 TODO: Получаем баланс начисленных, списанных баллов
	// 4.1.2 TODO: Если начисленных баллов больше, чем списанных, то разрешаем спсание
	// 4.1.3 TODO: Добавляем запись с количеством списанных баллов
	balance, err := srv.DBConnector.TryWithdraw(tokenHeader, orderWithdraw.Order, orderWithdraw.Withdraw)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка списания средств", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(HTTPAnswer(balance.ResponseStatus))

}

// GET
// 5 TODO: Получение списка ордеров по токену
func (srv *Server) apiUserOrdersGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenHeader := ""
	if r.Header["Authorization"] != nil {
		tokenHeader = r.Header["Authorization"][0]
	}
	// 5.1 TODO: Получение списка ордеров по токену в БД
	// 5.1.1 TODO: Из токена получаем имя пользователя
	// 5.1.2 TODO: По имени пользователя получаем ордера
	orders, err := srv.DBConnector.ListOrder(tokenHeader, srv.AddressAcSys)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	strJSON, err := orders.InJSON()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 5.2 TODO: Вывод ответа
	w.WriteHeader(HTTPAnswer(orders.ResponseStatus))
	_, err = w.Write(strJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 7 TODO: получаем баланс пользователя
func (srv *Server) apiUserBalanceGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenHeader := ""
	if r.Header["Authorization"] != nil {
		tokenHeader = r.Header["Authorization"][0]
	}

	// 7.1 TODO: Получаем баланс пользователя
	// 7.1 TODO: По токену получаем пользователя
	// 7.2 TODO: По пользовотелю получаем общий баланс начисленных и списанных баллов
	balances, err := srv.DBConnector.BalancesOrders(tokenHeader, srv.AddressAcSys)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка на сервере", http.StatusInternalServerError)
	}
	// 7.3 TODO: Вывод ответа
	strJSON, err := balances.InJSON()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(HTTPAnswer(balances.ResponseStatus))
	_, err = w.Write(strJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 8 TODO: Получение информации о выводе средств
func (srv *Server) apiUserWithdrawalsGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenHeader := ""
	if r.Header["Authorization"] != nil {
		tokenHeader = r.Header["Authorization"][0]
	}

	// 8.1 TODO: Получение информации о выводе средств в разрезе ордера
	withdraws, err := srv.DBConnector.UserWithdrawal(tokenHeader)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка на сервере", http.StatusInternalServerError)
	}

	// 8.2 TODO: Вывод ответа
	strJSON, err := withdraws.InJSON()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(HTTPAnswer(withdraws.ResponseStatus))
	_, err = w.Write(strJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 9 TODO: Взаимодействие с системой расчёта начислений баллов лояльности
func (srv *Server) apiUserAccrualGET(w http.ResponseWriter, r *http.Request) {
	number := mux.Vars(r)["number"]

	data := make(chan *postgresql.FullScoringSystem)
	// 9.1 TODO: Запускаем горутину с номером и каналом, где будет хранится ответ черного ящика
	// 9.1.1 TODO: Горутина запрашивает ответ от черного ящика.
	// 9.1.2 TODO: Если статус ответа не 429, то в канал пишется ответ горутина заканчивает свою работу
	// 9.1.3 TODO: Если статус ответа 429, то горутина засыпает на секунду и повторяет запрос к черному ящику
	// 9.1.3.1 TODO: так крутится пока не будет статус не 429
	go srv.ScoringSystem(number, data)

	// 9.2 TODO: Добавляет данные в БД. Вечный цикл с прослушиванием канала.
	// 9.2.1 TODO: Если в канале есть данные, то в БД добавляется запись начисления баллов ллояльности
	// 9.2.2 TODO: Если запись с начисление по ордеру есть в базе, то вторая запись не происходит
	fullScoringSystem := srv.executFSS(data)
	close(data)

	listAccrualJSON, err := json.MarshalIndent(fullScoringSystem.ScoringSystem, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	w.WriteHeader(HTTPAnswer(fullScoringSystem.ResponseStatus))
	_, err = w.Write(listAccrualJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) executFSS(data chan *postgresql.FullScoringSystem) (fullScoringSystem *postgresql.FullScoringSystem) {
	for {
		select {
		case <-data:
			fullScoringSystem := <-data
			_ = srv.SetValueScoringSystem(fullScoringSystem)
			return fullScoringSystem
		default:
			fmt.Println(0)
		}
	}
}
