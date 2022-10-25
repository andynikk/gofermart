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
	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	// 1.1 TODO: Проверка на gzip
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(r.Body)
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

	respByte, err := io.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
	}

	newAccount := new(postgresql.Account)
	newAccount.Cfg = new(postgresql.Cfg)

	newAccount.Pool = srv.Pool
	if err := json.Unmarshal(respByte, &newAccount.User); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
	}
	newAccount.Key = srv.Cfg.Key

	ctx, cancelFunc := context.WithCancel(context.Background())

	tx, err := srv.Pool.Begin(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		w.WriteHeader(http.StatusInternalServerError)
		cancelFunc = nil
		return
	}
	// 1.2 TODO: Регистрация пользователя в БД.
	// 1.2.1 TODO: Ищем пользовотеля в таблице БД. Если находим, то не создаем. Пароль кэшируется
	rez := newAccount.NewAccount()

	// 1.3 TODO: Создание токена
	tokenString := ""
	if rez == http.StatusOK {
		if err := tx.Commit(ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}

		// 1.3.1 TODO: Если пользователь добавлен создаем токен
		ct := token.Claims{Authorized: true, User: newAccount.Name, Exp: constants.TimeLiveToken}
		if tokenString, err = ct.GenerateJWT(); err != nil {
			constants.Logger.ErrorLog(err)
		}
	} else {
		if err := tx.Rollback(ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}
		w.WriteHeader(rez)
		cancelFunc = nil
		return
	}

	// 1.4 TODO: Добавление токена в Header
	w.Header().Add("Authorization", tokenString)
	w.WriteHeader(rez)

	cancelFunc()
}

// 2 TODO: Аутентификации пользователя
func (srv *Server) apiUserLoginPOST(w http.ResponseWriter, r *http.Request) {

	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	// 1.1 TODO: Проверка на gzip
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(r.Body)
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

	respByte, err := io.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка чтения тела", http.StatusInternalServerError)
	}

	Account := new(postgresql.Account)
	Account.Cfg = new(postgresql.Cfg)

	Account.Pool = srv.Pool
	if err := json.Unmarshal(respByte, &Account.User); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка Unmarshal", http.StatusInternalServerError)
	}
	Account.Key = srv.Cfg.Key

	// 2.1 TODO: Аутентификации пользователя в БД
	rez := Account.GetAccount()

	if rez != http.StatusOK {
		w.WriteHeader(rez)
		return
	}

	// 2.2 TODO: Создание токена
	tokenString := ""
	tc := token.Claims{Authorized: true, User: Account.Name, Exp: constants.TimeLiveToken}
	if tokenString, err = tc.GenerateJWT(); err != nil {
		constants.Logger.ErrorLog(err)
	}

	// 2.2 TODO: Добавление токена в Header
	w.Header().Add("Authorization", tokenString)
	//w.Header().Add("Set-Cookie", tokenString)
	w.WriteHeader(rez)
}

// 3 TODO: Добавление нового ордера
func (srv *Server) apiUserOrdersPOST(w http.ResponseWriter, r *http.Request) {
	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(r.Body)
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

	respByte, err := io.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Error get value", http.StatusInternalServerError)
	}

	numOrder, err := strconv.Atoi(string(respByte))
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "not Luna", http.StatusInternalServerError)
	}

	//TODO: Проверка на Луна
	if !luhn.Valid(numOrder) {
		constants.Logger.ErrorLog(err)
		http.Error(w, "not Luna", http.StatusUnprocessableEntity)
		return
	}

	order := new(postgresql.Order)
	order.Cfg = new(postgresql.Cfg)

	order.Number = strconv.Itoa(numOrder)
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	tx, err := srv.Pool.Begin(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		w.WriteHeader(http.StatusInternalServerError)
		cancelFunc = nil
		return
	}

	// 3.1 TODO: Добавление нового ордера в БД.
	// 3.1.1 TODO: Ищем ордер по номеру. Если не находим, то создаем
	rez := order.NewOrder()
	w.WriteHeader(rez)
	if rez == http.StatusOK {
		if err := tx.Commit(ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}
	} else {
		if err := tx.Rollback(ctx); err != nil {
			constants.Logger.ErrorLog(err)
		}
	}

	cancelFunc()
}

// 4 TODO: Списание баллов лояльности
func (srv *Server) apiUserWithdrawPOST(w http.ResponseWriter, r *http.Request) {
	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := r.Header.Get("Content-Encoding")

	bodyJSON = r.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(r.Body)
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

	respByte, err := io.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка чтения тела", http.StatusInternalServerError)
	}

	orderWithdraw := new(postgresql.OrderWithdraw)
	orderWithdraw.Cfg = new(postgresql.Cfg)

	orderWithdraw.Pool = srv.Pool
	if err := json.Unmarshal(respByte, &orderWithdraw); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(w, "Ошибка Unmarshal", http.StatusInternalServerError)
		return
	}
	if r.Header["Authorization"] != nil {
		orderWithdraw.Token = r.Header["Authorization"][0]
	}

	// 4.1 TODO: Списание баллов лояльности в БД
	// 4.1.1 TODO: Получаем баланс начисленных, списанных баллов
	// 4.1.2 TODO: Если начисленных баллов больше, чем списанных, то разрешаем спсание
	// 4.1.3 TODO: Добавляем запись с количеством списанных баллов
	w.WriteHeader(orderWithdraw.TryWithdraw())
}

// GET
// 5 TODO: Получение списка ордеров по токену
func (srv *Server) apiUserOrdersGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Cfg = new(postgresql.Cfg)
	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}
	// 5.1 TODO: Получение списка ордеров по токену в БД
	// 5.1.1 TODO: Из токена получаем имя пользователя
	// 5.1.2 TODO: По имени пользователя получаем ордера
	listOrder, status := order.ListOrder()

	if status != http.StatusOK {
		_, err := w.Write([]byte(""))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}

		w.WriteHeader(status)
		return
	}

	listOrderJSON, err := json.MarshalIndent(listOrder, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 5.1.4 TODO: Выводим список ордеров
	_, err = w.Write(listOrderJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	w.WriteHeader(status)
}

// 6 TODO: Для тестирования сделал API для продвижения ордера на следующий (рандомный) этап (/api/user/orders-next-status)
func (srv *Server) apiNextStatus(w http.ResponseWriter, r *http.Request) {

	order := new(postgresql.Order)
	order.Cfg = new(postgresql.Cfg)

	order.Pool = srv.Pool

	// 6.1 TODO: Двигаем ордер на следующий этап
	// 6.1.1 TODO: Получаем спсок ордеров по пользователю из токена
	// 6.1.2 TODO: Назначаем следующий этап. Если это статус PROCESSING, тогда выбираем рандомно INVALID или PROCESSED
	// 6.1.3 TODO: Устанавливаем статус и текущую дату соответствующей колонки
	// 6.1.4 TODO: Если это финальный этап (PROCESSED), рассчитываем баллы лояльности. Рандомное число между 100.10 и 501.98
	// 6.1.5 TODO: Добавляем баллы в ДБ
	order.SetNextStatus()

	w.WriteHeader(http.StatusOK)

	if err := r.Body.Close(); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

// 7 TODO: получаем баланс пользователя
func (srv *Server) apiUserBalanceGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Cfg = new(postgresql.Cfg)

	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	// 7.1 TODO: Получаем баланс пользователя
	// 7.1 TODO: По токену получаем пользователя
	// 7.2 TODO: По пользовотелю получаем общий баланс начисленных и списанных баллов
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

	// 7.3 TODO: Выводим в формате JSON
	w.WriteHeader(status)
	_, err = w.Write(listBalansJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
}

// 8 TODO: Получение информации о выводе средств
func (srv *Server) apiUserWithdrawalsGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	order := new(postgresql.Order)
	order.Cfg = new(postgresql.Cfg)

	order.Pool = srv.Pool
	if r.Header["Authorization"] != nil {
		order.Token = r.Header["Authorization"][0]
	}

	// 8.1 TODO: Получение информации о выводе средств в разрезе ордера
	listBalans, status := order.UserWithdrawal()
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, err := w.Write([]byte(""))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}
		return
	}

	// 8.2 TODO: Упаковка овета в JSON
	listBalansJSON, err := json.MarshalIndent(listBalans, "", " ")
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	w.WriteHeader(status)
	// 8.2 TODO: Вывод овета в JSON
	_, err = w.Write(listBalansJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
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

	w.WriteHeader(fullScoringSystem.HTTPStatus)
}

func (srv *Server) executFSS(data chan *postgresql.FullScoringSystem) (fullScoringSystem *postgresql.FullScoringSystem) {
	for {
		select {
		case <-data:
			fullScoringSystem := <-data
			srv.SetValueScoringSystem(fullScoringSystem)
			return fullScoringSystem

			//fullScoringSystem := <-data
			//srv.SetValueScoringSystem(<-data)
			//return <-data
			//return fullScoringSystem
		default:
			fmt.Println(0)
		}
	}
}
