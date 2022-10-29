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
	Address      string `env:"ADDRESS" envDefault:"localhost:8080"`
	AcSysAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"localhost:8000"`
}

type ServerConfig struct {
	Address      string
	AddressAcSys string
}

func (sc *ServerConfig) SetConfigServer() {

	addressPtr := flag.String("a", constants.PortServer, "порт сервера")
	addressAcSysPtr := flag.String("r", constants.PortAcSysServer, "сервер системы балов")
	flag.Parse()

	var cfgENV ServerConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	adresServer := cfgENV.Address
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		adresServer = *addressPtr
	}

	addressAcSysServer := cfgENV.AcSysAddress
	if _, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); !ok {
		addressAcSysServer = *addressAcSysPtr
	}

	sc.Address = adresServer
	sc.AddressAcSys = addressAcSysServer
	fmt.Println("---------------", sc.AddressAcSys)

}
