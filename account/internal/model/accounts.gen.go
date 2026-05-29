package model

import "time"

type Account struct {
	ID                   int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true"`
	CustomerID           int64     `gorm:"column:customer_id;type:bigint;not null;index:idx_accounts_customer_id"`
	Type                 int16     `gorm:"column:type;type:smallint;not null"`
	Status               int16     `gorm:"column:status;type:smallint;not null;default:1"`
	Currency             string    `gorm:"column:currency;type:character varying(8);not null"`
	CachedBalanceMinor   int64     `gorm:"column:cached_balance_minor;type:bigint;not null;default:0"`
	OverdraftLimitMinor  int64     `gorm:"column:overdraft_limit_minor;type:bigint;not null;default:0"`
	CreatedAt            time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	UpdatedAt            time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime"`
}
