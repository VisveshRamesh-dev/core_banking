package data

import (
	"fmt"

	"ledger/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewLedgerRepo)

type Data struct {
	db *gorm.DB
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(logger)

	db := c.GetDatabase()
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		db.GetHost(), db.GetPort(), db.GetUser(),
		db.GetPassword(), db.GetDbname(),
		db.GetSslmode(), db.GetTimezone(),
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		sqlDB, err := gormDB.DB()
		if err != nil {
			helper.Error(err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			helper.Error(err)
		}
		helper.Info("database connection closed")
	}

	helper.Info("database connection opened")
	return &Data{db: gormDB}, cleanup, nil
}
