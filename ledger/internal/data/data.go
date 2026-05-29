package data

import (
	"ledger/internal/conf"
	utilsdb "utils/db"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewLedgerRepo)

type Data struct {
	db *gorm.DB
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	db := c.GetDatabase()
	gormDB, cleanup, err := utilsdb.New(utilsdb.Config{
		Host:     db.GetHost(),
		Port:     db.GetPort(),
		User:     db.GetUser(),
		Password: db.GetPassword(),
		DBName:   db.GetDbname(),
		SSLMode:  db.GetSslmode(),
		TimeZone: db.GetTimezone(),
	}, logger)
	if err != nil {
		return nil, nil, err
	}
	return &Data{db: gormDB}, cleanup, nil
}
