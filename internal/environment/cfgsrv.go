package environment

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"

	"gofermart/internal/constants"
)

type ServerConfigENV struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

type ServerConfig struct {
	Address string
}

func (sc *ServerConfig) SetConfigServer() {

	addressPtr := flag.String("a", constants.PortServer, "порт сервера")
	flag.Parse()

	var cfgENV ServerConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	databaseDsn := cfgENV.Address
	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok {
		databaseDsn = *addressPtr
	}

	sc.Address = databaseDsn
}
