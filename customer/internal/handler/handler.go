package handler

import (
	"context"

	"customer/internal/data"

	"github.com/google/wire"
)

// customerRepo is the minimal interface the handler needs from the data layer.
// *data.CustomerRepo satisfies it; tests inject a mock.
type customerRepo interface {
	Create(ctx context.Context, rec *data.CustomerRecord) error
	GetByID(ctx context.Context, id int64) (*data.CustomerRecord, error)
	UpdateKYCStatus(ctx context.Context, id int64, newStatus int16) (*data.CustomerRecord, error)
	List(ctx context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error)
}

type CustomerHandler struct {
	customerRepo customerRepo
}

func NewCustomerHandler(repo *data.CustomerRepo) *CustomerHandler {
	return &CustomerHandler{customerRepo: repo}
}

var ProviderSet = wire.NewSet(NewCustomerHandler)
