package handler

import (
	"context"

	"transaction/internal/conf"
	"transaction/internal/data"
	"transaction/internal/model"

	"github.com/google/wire"
)

// transactionRepo is the minimal interface the handler needs.
type transactionRepo interface {
	Create(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id int64) (*model.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
	MarkCompleted(ctx context.Context, id int64, ledgerTxID int64) error
	MarkFailed(ctx context.Context, id int64, reason string) error
	ListByAccount(ctx context.Context, accountID int64, p data.ListParams) ([]*model.Transaction, int64, error)
}

type TransactionHandler struct {
	repo             transactionRepo
	accounts         data.AccountClient
	ledger           data.LedgerClient
	settlementAcctID int64
}

func NewTransactionHandler(
	repo *data.TransactionRepo,
	accounts data.AccountClient,
	ledger data.LedgerClient,
	app *conf.AppConf,
) *TransactionHandler {
	return &TransactionHandler{
		repo:             repo,
		accounts:         accounts,
		ledger:           ledger,
		settlementAcctID: app.SettlementAccountID,
	}
}

var ProviderSet = wire.NewSet(NewTransactionHandler)
