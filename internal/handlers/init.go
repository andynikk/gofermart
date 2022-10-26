package handlers

import (
	"github.com/gorilla/mux"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
	"github.com/andynikk/gofermart/internal/postgresql"
	"github.com/andynikk/gofermart/internal/token"
)

// 1 TODO: инициализация роутера и хендлеров
func (srv *Server) initRouters() {
	r := mux.NewRouter()

	//GET
	r.Handle("/api/user/orders", token.IsAuthorized(srv.apiUserOrdersGET)).Methods("GET")
	r.Handle("/api/user/balance", token.IsAuthorized(srv.apiUserBalanceGET)).Methods("GET")
	r.Handle("/api/user/balance/withdrawals", token.IsAuthorized(srv.apiUserWithdrawalsGET)).Methods("GET")
	r.Handle("/api/user/withdrawals", token.IsAuthorized(srv.apiUserWithdrawalsGET)).Methods("GET")

	r.Handle("/api/orders/{number}", token.IsAuthorized(srv.apiUserAccrualGET)).Methods("GET")

	r.HandleFunc("/api/user/orders-next-status", srv.apiNextStatus).Methods("GET")

	//POST
	r.Handle("/api/user/orders", token.IsAuthorized(srv.apiUserOrdersPOST)).Methods("POST")
	r.Handle("/api/user/balance/withdraw", token.IsAuthorized(srv.apiUserWithdrawPOST)).Methods("POST")

	//POST Handle Func
	r.HandleFunc("/api/user/register", srv.apiUserRegisterPOST).Methods("POST")
	r.HandleFunc("/api/user/login", srv.apiUserLoginPOST).Methods("POST")

	//GET HTTP PAGES
	r.HandleFunc("/", srv.HandleFunc).Methods("GET")
	r.HandleFunc("/user/register", srv.UserRegisterGET).Methods("GET")
	r.HandleFunc("/user/login", srv.UserLoginGET).Methods("GET")
	r.HandleFunc("/user/order", srv.UserOrderGET).Methods("GET")
	r.HandleFunc("/user/orders", srv.UserOrdersGET).Methods("GET")
	r.HandleFunc("/user/balance", srv.UserBalanceGET).Methods("GET")
	r.HandleFunc("/user/balance", srv.UserBalanceGET).Methods("GET")
	r.HandleFunc("/user/balance/withdraw", srv.UserBalanceWithdrawGET).Methods("GET")

	srv.Router = r
}

// 2 TODO: инициализация базы данных
func (srv *Server) InitDB() {
	srv.DBConnector = new(postgresql.DBConnector)
	if err := srv.DBConnector.PoolDB(); err != nil {
		constants.Logger.ErrorLog(err)
	}
	postgresql.CreateModeLDB(srv.Pool)
}

// 3 TODO: инициализация конфигурации
func (srv *Server) InitCFG() {
	srv.ServerConfig = new(environment.ServerConfig)
	srv.ServerConfig.SetConfigServer()

}
