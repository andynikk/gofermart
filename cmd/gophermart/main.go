package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/handlers"
)

type server struct {
	handlers.Server
}

func main() {
	server := new(server)
	handlers.NewServer(&server.Server)
	go func() {
		s := &http.Server{
			Addr:    server.Address,
			Handler: server.Router}

		if err := s.ListenAndServe(); err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
}
