package handler

import (
	"context"

	"ledger/internal/data"

	"github.com/google/wire"
)

// ledgerRepo is the minimal interface the handler needs.
// *data.LedgerRepo satisfies it; tests inject a mock.
type ledgerRepo interface {
	Post(ctx context.Context, tx *data.TxRecord) error
	GetByID(ctx context.Context, id int64) (*data.TxRecord, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*data.TxRecord, error)
	GetBalance(ctx context.Context, accountID int64) (*data.BalanceResult, error)
}

type LedgerHandler struct {
	repo ledgerRepo
}

func NewLedgerHandler(repo *data.LedgerRepo) *LedgerHandler {
	return &LedgerHandler{repo: repo}
}

var ProviderSet = wire.NewSet(NewLedgerHandler)
