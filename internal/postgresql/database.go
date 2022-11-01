package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/environment"
)

type DBConnector struct {
	Pool *pgxpool.Pool
	Cfg  *environment.DBConfig
}

func (dbc *DBConnector) PoolDB() error {
	dbCfg := new(environment.DBConfig)
	dbCfg.SetConfigDB()

	if dbCfg.DatabaseDsn == "" {
		return errors.New("пустой путь к базе")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	pool, err := pgxpool.Connect(ctx, dbCfg.DatabaseDsn)
	if err != nil {
		cancelFunc = nil
		return err
	}

	dbc.Pool = pool
	dbc.Cfg = dbCfg

	cancelFunc()
	return nil

}
