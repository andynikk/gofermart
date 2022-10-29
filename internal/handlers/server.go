package handlers

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
	"github.com/andynikk/gofermart/internal/postgresql"
	"github.com/andynikk/gofermart/internal/token"
)

type Server struct {
	*mux.Router
	*postgresql.DBConnector
	*environment.ServerConfig
}

func NewByConfig() (srv *Server) {
	srv = new(Server)
	srv.initRouters()
	srv.InitDB()
	srv.InitCFG()
	srv.InitScoringSystem()

	return srv
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

	rw.WriteHeader(http.StatusOK)
}

func (srv *Server) HandlerNotFound(rw http.ResponseWriter, r *http.Request) {

	http.Error(rw, "Page "+r.URL.Path+" not found", http.StatusNotFound)

}

func (srv *Server) Shutdown() {
	constants.Logger.InfoLog("server stopped")
}

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
	r.HandleFunc("/docs/user/register", srv.UserRegisterGET).Methods("GET")
	r.HandleFunc("/docs/user/login", srv.UserLoginGET).Methods("GET")
	r.HandleFunc("/docs/user/order", srv.UserOrderGET).Methods("GET")
	r.HandleFunc("/docs/user/orders", srv.UserOrdersGET).Methods("GET")
	r.HandleFunc("/docs/user/balance", srv.UserBalanceGET).Methods("GET")
	r.HandleFunc("/docs/user/balance", srv.UserBalanceGET).Methods("GET")
	r.HandleFunc("/docs/user/balance/withdraw", srv.UserBalanceWithdrawGET).Methods("GET")
	r.HandleFunc("/docs/user/balance/withdrawals", srv.UserBalanceWithdrawsGET).Methods("GET")
	r.HandleFunc("/docs/user/accrual", srv.UserAccrualGET).Methods("GET")

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

// 4 TODO: инициализация системы лояльности
func (srv *Server) InitScoringSystem() {
	//good := Goods{
	//	"My table",
	//	15,
	//	"%",
	//}
	//srv.AddItemsScoringSystem(&good)
	//
	//good = Goods{
	//	"You table",
	//	10,
	//	"%",
	//}
	//srv.AddItemsScoringSystem(&good)
}

func HTTPAnswer(answer constants.Answer) int {

	HTTPAnswer := 0
	switch answer {
	case constants.AnswerSuccessfully:
		HTTPAnswer = http.StatusOK //200

	case constants.AnswerInvalidFormat:
		HTTPAnswer = http.StatusBadRequest //400

	case constants.AnswerLoginBusy:
		HTTPAnswer = http.StatusConflict //409

	case constants.AnswerErrorServer:
		HTTPAnswer = http.StatusInternalServerError //500

	case constants.AnswerInvalidLoginPassword:
		HTTPAnswer = http.StatusUnauthorized //401

	case constants.AnswerUserNotAuthenticated:
		HTTPAnswer = http.StatusUnauthorized //401

	case constants.AnswerAccepted:
		HTTPAnswer = http.StatusAccepted //202

	case constants.AnswerUploadedAnotherUser:
		HTTPAnswer = http.StatusConflict //409

	case constants.AnswerInvalidOrderNumber:
		HTTPAnswer = http.StatusUnprocessableEntity //422

	case constants.AnswerInsufficientFunds:
		HTTPAnswer = http.StatusPaymentRequired //402

	case constants.AnswerNoContent:
		HTTPAnswer = http.StatusNoContent //204

	case constants.AnswerConflict:
		HTTPAnswer = http.StatusConflict //409

	case constants.AnswerTooManyRequests:
		HTTPAnswer = http.StatusTooManyRequests //429
	default:
		HTTPAnswer = 0
	}

	return HTTPAnswer
}
