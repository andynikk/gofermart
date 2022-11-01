package main

import (
	"github.com/andynikk/gofermart/internal/handlers"
)

// TODO: запуск сервера
func main() {
	srv := handlers.NewByConfig()
	srv.Run()
}
