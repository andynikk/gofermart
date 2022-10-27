package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/gofermart/internal/handlers"
)

// TODO: запуск сервера
func main() {
	server := handlers.NewServer()
	go func() {
		s := &http.Server{
			Addr:    server.Address,
			Handler: server.Router}

		if err := s.ListenAndServe(); err != nil {
			log.Fatalln(err.Error())
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop

	server.Shutdown()
}
