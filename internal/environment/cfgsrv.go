package environment

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/andynikk/gofermart/internal/channel"
	"github.com/andynikk/gofermart/internal/constants"
)

type ServerConfigENV struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8000"`
	DemoMode       string `env:"DEMO_MODE" envDefault:"0"`
}

type ServerConfig struct {
	Address        string
	AccrualAddress string
	DemoMode       bool
	ChanData       chan *channel.ScoringOrder
}

func NewConfigServer() (*ServerConfig, error) {

	addressPtr := flag.String("a", constants.PortServer, "порт сервера")
	addressAcSysPtr := flag.String("r", constants.PortAcSysServer, "сервер системы балов")
	demoMode := flag.Bool("m", constants.DemoMode, "это демо режим")
	flag.Parse()

	var cfgENV ServerConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	addresServer := cfgENV.Address
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		addresServer = *addressPtr
	}

	addressAcSysServer := cfgENV.AccrualAddress
	if _, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); !ok {
		addressAcSysServer = *addressAcSysPtr
	}

	demoModeServer := false
	if cfgENV.DemoMode != "0" {
		demoModeServer = true
	}
	if _, ok := os.LookupEnv("DEMO_MODE"); !ok {
		demoModeServer = *demoMode
	}

	sc := ServerConfig{
		addresServer,
		addressAcSysServer,
		demoModeServer,
		make(chan *channel.ScoringOrder),
	}
	return &sc, err
}
