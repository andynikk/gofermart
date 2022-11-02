package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/theplant/luhn"

	"github.com/andynikk/gofermart/internal/channel"
	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/postgresql"
	"github.com/andynikk/gofermart/internal/token"
)

func (srv *Server) HandlerNotFound(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Page "+r.URL.Path+" not found", http.StatusNotFound)
}

func (srv *Server) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	content := srv.StartPage()

	acceptEncoding := rq.Header.Get("Accept-Encoding")

	metricsHTML := []byte(content)
	byteMterics := bytes.NewBuffer(metricsHTML).Bytes()
	compData, err := compression.Compress(byteMterics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	var bodyBate []byte
	if strings.Contains(acceptEncoding, "gzip") {
		rw.Header().Add("Content-Encoding", "gzip")
		bodyBate = compData
	} else {
		bodyBate = metricsHTML
	}

	rw.Header().Add("Content-Type", "text/html")
	if _, err := rw.Write(bodyBate); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	//rw.WriteHeader(http.StatusOK)
}

// POST
// 1 TODO: Регистрация пользователя
func (srv *Server) apiUserRegisterPOST(w http.ResponseWriter, r *http.Request) {

	user := postgresql.User{}
	if err := user.Unmarshal([]byte(r.Header.Get(constants.HeaderMiddlewareBody))); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 1.2 TODO: Регистрация пользователя в БД.
	// 1.2.1 TODO: Ищем пользовотеля в таблице БД. Если находим, то не создаем. Пароль кэшируется
	tokenString := ""
	account, err := srv.DBConnector.NewAccount(user.Name, user.Password)
	if err != nil {
		w.Header().Add(constants.HeaderAuthorization, tokenString)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}

	// 1.3 TODO: Создание токена
	// 1.3.1 TODO: Если пользователь добавлен создаем токен
	tc := token.NewClaims(account.Name)
	if tokenString, err = tc.GenerateJWT(); err != nil {
		w.Header().Add(constants.HeaderAuthorization, "")
		http.Error(w, "Ошибка получения токена", http.StatusInternalServerError)
		return
	}

	// 1.4 TODO: Добавление токена в Header
	w.Header().Add(constants.HeaderAuthorization, tokenString)
	w.WriteHeader(http.StatusOK)
}

// 2 TODO: Аутентификации пользователя
func (srv *Server) apiUserLoginPOST(w http.ResponseWriter, r *http.Request) {

	user := postgresql.User{}

	if err := user.Unmarshal([]byte(r.Header.Get(constants.HeaderMiddlewareBody))); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 2.1 TODO: Аутентификации пользователя в БД
	tokenString := ""
	account, err := srv.DBConnector.GetAccount(user.Name, user.Password)
	if err != nil {
		w.Header().Add(constants.HeaderAuthorization, tokenString)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}

	// 2.2 TODO: Создание токена
	tc := token.NewClaims(account.Name)
	if tokenString, err = tc.GenerateJWT(); err != nil {
		w.Header().Add(constants.HeaderAuthorization, "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2.2 TODO: Добавление токена в Header
	w.Header().Add(constants.HeaderAuthorization, tokenString)
	w.WriteHeader(http.StatusOK)
}

// 3 TODO: Добавление нового ордера
func (srv *Server) apiUserOrdersPOST(w http.ResponseWriter, r *http.Request) {

	respByte := r.Header.Get(constants.HeaderMiddlewareBody)
	numOrder, err := strconv.Atoi(respByte)
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
	if r.Header[constants.HeaderAuthorization] != nil {
		tokenHeader = r.Header[constants.HeaderAuthorization][0]
	}

	// 3.1 TODO: Добавление нового ордера в БД.
	// 3.1.1 TODO: Ищем ордер по номеру. Если не находим, то создаем
	if _, err := srv.DBConnector.NewOrder(tokenHeader, numOrder); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}
	w.WriteHeader(http.StatusAccepted)

	if srv.DemoMode {

		goodOrderSS := NewGoodOrderSS()

		var arrGoodOrderSS []GoodOrderSS
		arrGoodOrderSS = append(arrGoodOrderSS, *goodOrderSS)

		orderSS := OrderSS{
			string(respByte),
			arrGoodOrderSS,
		}
		err = srv.AddOrderScoringOrder(&orderSS)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}

		if _, err = srv.DBConnector.SetStartedAt(numOrder, tokenHeader); err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// 4 TODO: Списание баллов лояльности
func (srv *Server) apiUserWithdrawPOST(w http.ResponseWriter, r *http.Request) {

	withdraw := postgresql.Withdraw{}
	if err := withdraw.Unmarshal([]byte(r.Header.Get(constants.HeaderMiddlewareBody))); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка Unmarshal", http.StatusInternalServerError)
		return
	}

	tokenHeader := ""
	if r.Header[constants.HeaderAuthorization] != nil {
		tokenHeader = r.Header[constants.HeaderAuthorization][0]
	}

	// 4.1 TODO: Списание баллов лояльности в БД
	// 4.1.1 TODO: Получаем баланс начисленных, списанных баллов
	// 4.1.2 TODO: Если начисленных баллов больше, чем списанных, то разрешаем спсание
	// 4.1.3 TODO: Добавляем запись с количеством списанных баллов
	if _, err := srv.DBConnector.TryWithdraw(tokenHeader, withdraw.Order, withdraw.Withdraw); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}
	w.WriteHeader(http.StatusOK)

}

// GET
// 5 TODO: Получение списка ордеров по токену
func (srv *Server) apiUserOrdersGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenHeader := ""
	if r.Header[constants.HeaderAuthorization] != nil {
		tokenHeader = r.Header[constants.HeaderAuthorization][0]
	}
	// 5.1 TODO: Получение списка ордеров по токену в БД
	// 5.1.1 TODO: Из токена получаем имя пользователя
	// 5.1.2 TODO: По имени пользователя получаем ордера
	orders, err := srv.DBConnector.ListOrder(tokenHeader, srv.AccrualAddress)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	strJSON, err := orders.Marshal()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 5.2 TODO: Вывод ответа
	w.WriteHeader(http.StatusOK)
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
	if r.Header[constants.HeaderAuthorization] != nil {
		tokenHeader = r.Header[constants.HeaderAuthorization][0]
	}

	// 7.1 TODO: Получаем баланс пользователя
	// 7.1 TODO: По токену получаем пользователя
	// 7.2 TODO: По пользовотелю получаем общий баланс начисленных и списанных баллов
	balances, err := srv.DBConnector.BalancesOrders(tokenHeader, srv.AccrualAddress)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}
	// 7.3 TODO: Вывод ответа
	strJSON, err := balances.Marshal()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	if r.Header[constants.HeaderAuthorization] != nil {
		tokenHeader = r.Header[constants.HeaderAuthorization][0]
	}

	// 8.1 TODO: Получение информации о выводе средств в разрезе ордера
	withdraws, err := srv.DBConnector.UserWithdrawal(tokenHeader)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}

	// 8.2 TODO: Вывод ответа
	strJSON, err := withdraws.Marshal()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(strJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// 9 TODO: Взаимодействие с системой расчёта начислений баллов лояльности
func (srv *Server) apiUserAccrualGET(w http.ResponseWriter, r *http.Request) {
	number := mux.Vars(r)["number"]

	//data := make(chan *postgresql.FullScoringOrder)
	// 9.1 TODO: Запускаем горутину с номером и каналом, где будет хранится ответ черного ящика
	// 9.1.1 TODO: Горутина запрашивает ответ от черного ящика.
	// 9.1.2 TODO: Если статус ответа не 429, то в канал пишется ответ горутина заканчивает свою работу
	// 9.1.3 TODO: Если статус ответа 429, то горутина засыпает на секунду и повторяет запрос к черному ящику
	// 9.1.3.1 TODO: так крутится пока не будет статус не 429
	go srv.ScoringOrder(number)

	// 9.2 TODO: Добавляет данные в БД. Вечный цикл с прослушиванием канала.
	// 9.2.1 TODO: Если в канале есть данные, то в БД добавляется запись начисления баллов ллояльности
	// 9.2.2 TODO: Если запись с начисление по ордеру есть в базе, то вторая запись не происходит
	fullScoringOrder := srv.executFSS()

	listAccrualJSON, err := json.MarshalIndent(fullScoringOrder, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, err.Error(), HTTPErrors(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(listAccrualJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (srv *Server) executFSS() (scoringOrder *channel.ScoringOrder) {
	for {
		select {
		case <-srv.ChanData:
			scoringOrder := <-srv.ChanData
			_ = srv.SetValueScoringOrder(scoringOrder)
			return scoringOrder
		default:
			fmt.Println(0)
		}
	}
}

func (srv *Server) Shutdown() {
	constants.Logger.InfoLog("server stopped")
}
