package environment

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/andynikk/gofermart/internal/constants"
)

type DBConfig struct {
	DatabaseDsn string
	Key         string
}

type DBConfigENV struct {
	DatabaseDsn string `env:"DATABASE_DSN"`
	Key         string `env:"KEY"`
}

func (dbc *DBConfig) SetConfigDB() {

	keyDatabaseDsn := flag.String("d", "", "строка соединения с базой")
	keyFlag := flag.String("k", "", "ключ хеша")
	flag.Parse()

	var cfgENV DBConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	databaseDsn := cfgENV.DatabaseDsn
	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok {
		databaseDsn = *keyDatabaseDsn
	}

	keyHash := cfgENV.Key
	if _, ok := os.LookupEnv("KEY"); !ok {
		keyHash = *keyFlag
	}
	if keyHash == "" {
		keyHash = string(constants.HashKey[:])
	}

	dbc.DatabaseDsn = databaseDsn
	dbc.Key = keyHash
}
