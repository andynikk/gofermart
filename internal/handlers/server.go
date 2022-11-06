package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
	"github.com/andynikk/gofermart/internal/middlware"
	"github.com/andynikk/gofermart/internal/postgresql"
)

type IServer interface {
	initRouters()
	initDataBase()
	initConfig()
	initScoringSystem()
	Run()
}

func NewByConfig() (s IServer) {
	var srv IServer = &Server{}

	srv.initRouters()
	srv.initDataBase()
	srv.initConfig()
	srv.initScoringSystem()

	return srv
}

type Server struct {
	*mux.Router
	*postgresql.DBConnector
	*environment.ServerConfig
}

func (srv *Server) Run() {
	go func() {
		s := &http.Server{
			Addr:    srv.Address,
			Handler: srv.Router}

		if err := s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
	srv.Shutdown()
}

// 1 TODO: инициализация роутера и хендлеров
func (srv *Server) initRouters() {
	r := mux.NewRouter()
	r.Use(middlware.GzipMiddlware)

	//GET
	r.Handle("/api/user/orders", middlware.IsAuthorized(srv.apiUserOrdersGET)).Methods("GET")
	r.Handle("/api/user/balance", middlware.IsAuthorized(srv.apiUserBalanceGET)).Methods("GET")
	r.Handle("/api/user/balance/withdrawals", middlware.IsAuthorized(srv.apiUserWithdrawalsGET)).Methods("GET")
	r.Handle("/api/user/withdrawals", middlware.IsAuthorized(srv.apiUserWithdrawalsGET)).Methods("GET")

	r.Handle("/api/orders/{number}", middlware.IsAuthorized(srv.apiUserAccrualGET)).Methods("GET")

	//r.HandleFunc("/api/user/orders-next-status", srv.apiNextStatus).Methods("GET")

	//POST
	r.Handle("/api/user/orders", middlware.IsAuthorized(srv.apiUserOrdersPOST)).Methods("POST")
	r.Handle("/api/user/balance/withdraw", middlware.IsAuthorized(srv.apiUserWithdrawPOST)).Methods("POST")

	//POST Handle Func
	r.HandleFunc("/api/user/register", srv.apiUserRegisterPOST).Methods("POST")
	r.HandleFunc("/api/user/login", srv.apiUserLoginPOST).Methods("POST")

	//GET HTTP PAGES
	r.HandleFunc("/", srv.HandleFunc).Methods("GET")
	r.HandleFunc("/docs/user/register", srv.UserRegisterGET).Methods("GET")
	r.HandleFunc("/docs/user/login", srv.UserLoginGET).Methods("GET")
	r.HandleFunc("/docs/user/order", srv.UserOrderGET).Methods("GET")
	r.HandleFunc("/docs/user/orders", srv.UserOrdersGET).Methods("GET")
	r.HandleFunc("/docs/user/balance", srv.UserBalanceGET).Methods("GET")
	r.HandleFunc("/docs/user/balance", srv.UserBalanceGET).Methods("GET")
	r.HandleFunc("/docs/user/balance/withdraw", srv.UserBalanceWithdrawGET).Methods("GET")
	r.HandleFunc("/docs/user/balance/withdrawals", srv.UserBalanceWithdrawsGET).Methods("GET")
	r.HandleFunc("/docs/user/accrual", srv.UserAccrualGET).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(srv.HandlerNotFound)

	srv.Router = r
}

// 2 TODO: инициализация базы данных
func (srv *Server) initDataBase() {
	dbc, err := postgresql.NewDBConnector()
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	srv.DBConnector = dbc
	postgresql.CreateModeLDB(srv.Pool)
}

// 3 TODO: инициализация конфигурации
func (srv *Server) initConfig() {
	srvConfig, err := environment.NewConfigServer()
	if err != nil {
		log.Fatal(err)
	}
	srv.ServerConfig = srvConfig

}

// 4 TODO: инициализация системы лояльности
func (srv *Server) initScoringSystem() {
	if !srv.DemoMode {
		return
	}

	good := Goods{
		"My table",
		15,
		"%",
	}
	srv.AddItemsScoringOrder(&good)

	good = Goods{
		"You table",
		10,
		"%",
	}
	srv.AddItemsScoringOrder(&good)
}

func HTTPErrors(err error) int {

	HTTPAnswer := http.StatusOK

	if errors.Is(err, constants.ErrInvalidFormat) {
		HTTPAnswer = http.StatusBadRequest //400
	} else if errors.Is(err, constants.ErrLoginBusy) {
		HTTPAnswer = http.StatusConflict //409
	} else if errors.Is(err, constants.ErrErrorServer) {
		HTTPAnswer = http.StatusInternalServerError //500
	} else if errors.Is(err, constants.ErrInvalidLoginPassword) {
		HTTPAnswer = http.StatusUnauthorized //401
	} else if errors.Is(err, constants.ErrUserNotAuthenticated) {
		HTTPAnswer = http.StatusUnauthorized //401
	} else if errors.Is(err, constants.ErrAccepted) {
		HTTPAnswer = http.StatusAccepted //202
	} else if errors.Is(err, constants.ErrUploadedAnotherUser) {
		HTTPAnswer = http.StatusConflict //409
	} else if errors.Is(err, constants.ErrInvalidOrderNumber) {
		HTTPAnswer = http.StatusUnprocessableEntity //422
	} else if errors.Is(err, constants.ErrInsufficientFunds) {
		HTTPAnswer = http.StatusPaymentRequired //402
	} else if errors.Is(err, constants.ErrNoContent) {
		HTTPAnswer = http.StatusNoContent //204
	} else if errors.Is(err, constants.ErrConflict) {
		HTTPAnswer = http.StatusConflict //409
	} else if errors.Is(err, constants.ErrTooManyRequests) {
		HTTPAnswer = http.StatusTooManyRequests //429
	} else if errors.Is(err, constants.ErrOrderUpload) {
		HTTPAnswer = http.StatusOK //200
	}
	return HTTPAnswer
}
