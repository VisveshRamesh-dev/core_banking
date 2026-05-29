package data

import (
	"context"
	"errors"
	"time"

	"transaction/internal/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

// ListParams carries pagination for listing transactions.
type ListParams struct {
	Limit  int
	Offset int
}

type TransactionRepo struct {
	db *gorm.DB
}

func NewTransactionRepo(data *Data) *TransactionRepo {
	return &TransactionRepo{db: data.db}
}

func (r *TransactionRepo) Create(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *TransactionRepo) GetByID(ctx context.Context, id int64) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.db.WithContext(ctx).First(&tx, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&tx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

// MarkCompleted transitions the transaction to COMPLETED and records the ledger tx ID.
func (r *TransactionRepo) MarkCompleted(ctx context.Context, id int64, ledgerTxID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"state":                  model.TxStateCompleted,
			"ledger_transaction_id":  ledgerTxID,
			"completed_at":           now,
		}).Error
}

// MarkFailed transitions the transaction to FAILED and records why.
func (r *TransactionRepo) MarkFailed(ctx context.Context, id int64, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"state":          model.TxStateFailed,
			"failure_reason": reason,
			"completed_at":   now,
		}).Error
}

// ListByAccount returns transactions where the account appears as sender or receiver.
func (r *TransactionRepo) ListByAccount(ctx context.Context, accountID int64, p ListParams) ([]*model.Transaction, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txs []model.Transaction
	if err := q.Order("id DESC").Limit(p.Limit).Offset(p.Offset).Find(&txs).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*model.Transaction, len(txs))
	for i := range txs {
		result[i] = &txs[i]
	}
	return result, total, nil
}
