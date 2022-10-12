package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gofermart/internal/constants"
	"gofermart/internal/handlers"
)

type server struct {
	handlers.Server
}

func main() {

	server := new(server)
	handlers.NewServer(&server.Server)

	fmt.Println("/////////////////", server.Address)
	go func() {
		fmt.Println("/////////////////", server.Address)
		s := &http.Server{
			Addr:    server.Address,
			Handler: server.Router}

		if err := s.ListenAndServe(); err != nil {
			fmt.Println("/////////////////", err)
			constants.Logger.ErrorLog(err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop

}
