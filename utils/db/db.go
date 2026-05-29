package db

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds the fields needed to open a PostgreSQL connection.
type Config struct {
	Host     string
	Port     int32
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// New opens a GORM/PostgreSQL connection and returns a cleanup function.
func New(c Config, logger log.Logger) (*gorm.DB, func(), error) {
	helper := log.NewHelper(logger)

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.TimeZone,
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
	return gormDB, cleanup, nil
}
