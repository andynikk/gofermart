package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/andynikk/gofermart/internal/compression"
	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/web"
)

func (srv *Server) UserRegisterGET(rw http.ResponseWriter, rq *http.Request) {
	content := web.RegisterPage()

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

func (srv *Server) UserLoginGET(rw http.ResponseWriter, rq *http.Request) {

	content := web.LoginPage()

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

func (srv *Server) AuthorizationPageGET(rw http.ResponseWriter, rq *http.Request) {
	content := web.LoginPage()

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
