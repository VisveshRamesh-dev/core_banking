package model

import "time"

type Transaction struct {
	ID                   int64      `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true"`
	Kind                 int16      `gorm:"column:kind;type:smallint;not null"`
	State                int16      `gorm:"column:state;type:smallint;not null;default:1"`
	FromAccountID        *int64     `gorm:"column:from_account_id;type:bigint"`
	ToAccountID          *int64     `gorm:"column:to_account_id;type:bigint"`
	AmountMinor          int64      `gorm:"column:amount_minor;type:bigint;not null"`
	Currency             string     `gorm:"column:currency;type:character varying(8);not null"`
	IdempotencyKey       string     `gorm:"column:idempotency_key;type:character varying(255);not null;uniqueIndex:idx_transactions_idempotency"`
	LedgerTransactionID  *int64     `gorm:"column:ledger_transaction_id;type:bigint"`
	FailureReason        *string    `gorm:"column:failure_reason;type:character varying(500)"`
	SourceReference      *string    `gorm:"column:source_reference;type:character varying(255)"`
	DestinationReference *string    `gorm:"column:destination_reference;type:character varying(255)"`
	CreatedAt            time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	CompletedAt          *time.Time `gorm:"column:completed_at;type:timestamptz"`
}
