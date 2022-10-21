package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
	"github.com/andynikk/gofermart/internal/postgresql"
	"github.com/andynikk/gofermart/internal/token"
	"github.com/andynikk/gofermart/internal/web"
)

//var mySigningKey = []byte("johenews")

type Server struct {
	*mux.Router
	*postgresql.DBConnector
	*environment.ServerConfig
}

func NewServer(srv *Server) {

	r := mux.NewRouter()

	//limiter := make(chan struct{}, 500)
	//r.Use(func(handler http.Handler) http.Handler {
	//	limiter <- struct{}{}
	//	defer func() { <-limiter }()
	//
	//	constants.Logger.InfoLog("Я тут")
	//	return handler
	//})

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

	srv.Router = r

	srv.DBConnector = new(postgresql.DBConnector)
	if err := srv.DBConnector.PoolDB(); err != nil {
		constants.Logger.ErrorLog(err)
	}

	srv.ServerConfig = new(environment.ServerConfig)
	srv.ServerConfig.SetConfigServer()

	postgresql.CreateModeLDB(srv.Pool)
}

func (srv *Server) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	content := web.StartPage()

	cookie, err := rq.Cookie(constants.AccountCookies)
	if err == nil {
		arrOrder := new([]string)
		content = web.OrderPage(*arrOrder)
		fmt.Println(cookie.Path)
	}

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

	rw.WriteHeader(http.StatusOK)
}

func (srv *Server) HandlerNotFound(rw http.ResponseWriter, r *http.Request) {

	http.Error(rw, "Page "+r.URL.Path+" not found", http.StatusNotFound)

}
