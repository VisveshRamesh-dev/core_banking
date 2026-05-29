package data

import (
	"context"
	"errors"
	"time"

	"ledger/internal/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

// TxRecord bundles a transaction with its entries.
type TxRecord struct {
	Tx      model.LedgerTransaction
	Entries []model.LedgerEntry
}

// BalanceResult holds the computed balance for an account.
type BalanceResult struct {
	Balance  int64
	Currency string
	AsOf     time.Time
}

type LedgerRepo struct {
	db *gorm.DB
}

func NewLedgerRepo(data *Data) *LedgerRepo {
	return &LedgerRepo{db: data.db}
}

// Post inserts a transaction and all its entries atomically.
// rec.Tx.ID and each entry's TransactionID are back-filled before returning.
func (r *LedgerRepo) Post(ctx context.Context, rec *TxRecord) error {
	return r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if err := db.Create(&rec.Tx).Error; err != nil {
			return err
		}
		for i := range rec.Entries {
			rec.Entries[i].TransactionID = rec.Tx.ID
		}
		return db.Create(&rec.Entries).Error
	})
}

// GetByID fetches a transaction and its entries by primary key.
func (r *LedgerRepo) GetByID(ctx context.Context, id int64) (*TxRecord, error) {
	var tx model.LedgerTransaction
	if err := r.db.WithContext(ctx).First(&tx, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.loadEntries(ctx, tx)
}

// GetByIdempotencyKey looks up a transaction by its idempotency key.
func (r *LedgerRepo) GetByIdempotencyKey(ctx context.Context, key string) (*TxRecord, error) {
	var tx model.LedgerTransaction
	if err := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&tx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.loadEntries(ctx, tx)
}

// GetBalance sums all POSTED entries for an account and returns the net balance.
func (r *LedgerRepo) GetBalance(ctx context.Context, accountID int64) (*BalanceResult, error) {
	type row struct {
		Balance  int64
		Currency string
		AsOf     time.Time
	}
	var result row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN e.direction = ? THEN e.amount_minor ELSE -e.amount_minor END), 0) AS balance,
			COALESCE(MAX(e.currency), '')                                                              AS currency,
			COALESCE(MAX(t.posted_at), NOW())                                                         AS as_of
		FROM ledger_entries e
		JOIN ledger_transactions t ON t.id = e.transaction_id
		WHERE e.account_id = ? AND t.status = ?
	`, model.DirectionCredit, accountID, model.TxStatusPosted).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &BalanceResult{
		Balance:  result.Balance,
		Currency: result.Currency,
		AsOf:     result.AsOf,
	}, nil
}

func (r *LedgerRepo) loadEntries(ctx context.Context, tx model.LedgerTransaction) (*TxRecord, error) {
	var entries []model.LedgerEntry
	if err := r.db.WithContext(ctx).
		Where("transaction_id = ?", tx.ID).
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return &TxRecord{Tx: tx, Entries: entries}, nil
}
