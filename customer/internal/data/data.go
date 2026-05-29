package data

import (
	"customer/internal/conf"
	utilsdb "utils/db"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCustomerRepo)

// Data holds the database handle shared across all repositories in this service.
type Data struct {
	db *gorm.DB
}

// NewData opens the PostgreSQL connection using the shared utils/db helper.
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
