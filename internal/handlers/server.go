package handlers

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
	"github.com/andynikk/gofermart/internal/postgresql"
	"github.com/andynikk/gofermart/internal/web"
)

type Server struct {
	*mux.Router
	*postgresql.DBConnector
	*environment.ServerConfig
}

func NewServer() (srv *Server) {
	srv = new(Server)
	srv.initRouters()
	srv.InitDB()
	srv.InitCFG()
	srv.InitScoringSystem()

	return srv
}

func (srv *Server) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	content := web.StartPage(srv.Address)

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
