package handler

import (
	"context"

	"account/internal/data"
	"account/internal/model"

	"github.com/google/wire"
)

// accountRepo is the minimal interface the handler needs; *data.AccountRepo satisfies it.
type accountRepo interface {
	Create(ctx context.Context, a *model.Account) error
	GetByID(ctx context.Context, id int64) (*model.Account, error)
	UpdateStatus(ctx context.Context, id int64, newStatus int16) (*model.Account, error)
	ListByCustomer(ctx context.Context, customerID int64, p data.ListParams) ([]*model.Account, int64, error)
}

type AccountHandler struct {
	repo      accountRepo
	customers data.CustomerClient
	ledger    data.LedgerClient
}

func NewAccountHandler(repo *data.AccountRepo, customers data.CustomerClient, ledger data.LedgerClient) *AccountHandler {
	return &AccountHandler{repo: repo, customers: customers, ledger: ledger}
}

var ProviderSet = wire.NewSet(NewAccountHandler)
