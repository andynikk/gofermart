package environment

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/andynikk/gofermart/internal/constants"
)

type ServerConfigENV struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

type ServerConfig struct {
	Address string
}

func (sc *ServerConfig) SetConfigServer() {

	fmt.Println("********1")
	addressPtr := flag.String("a", constants.PortServer, "порт сервера")
	flag.Parse()

	fmt.Println("********2")
	var cfgENV ServerConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("********3")
	databaseDsn := cfgENV.Address
	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok {
		databaseDsn = *addressPtr
	}

	sc.Address = databaseDsn
}
