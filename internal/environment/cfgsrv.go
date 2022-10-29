package environment

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/andynikk/gofermart/internal/constants"
)

type ServerConfigENV struct {
	Address      string `env:"ADDRESS" envDefault:"localhost:8080"`
	AcSysAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"localhost:8000"`
	DemoMode     bool   `env:"DEMO" envDefault:"false"`
}

type ServerConfig struct {
	Address      string
	AddressAcSys string
	DemoMode     bool
}

func (sc *ServerConfig) SetConfigServer() {

	addressPtr := flag.String("a", constants.PortServer, "порт сервера")
	addressAcSysPtr := flag.String("r", constants.PortAcSysServer, "сервер системы балов")
	demoMode := flag.Bool("m", constants.DemoMode, "это демо режим")
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

	demoModeServer := cfgENV.DemoMode
	if _, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); !ok {
		demoModeServer = *demoMode
	}

	sc.Address = adresServer
	sc.AddressAcSys = addressAcSysServer
	sc.DemoMode = demoModeServer

}
