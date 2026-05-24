package handler

import (
	"customer/internal/data"

	"github.com/google/wire"
)

type CustomerHandler struct {
	customerRepo *data.CustomerRepo
}

func NewCustomerHandler(customerRepo *data.CustomerRepo) *CustomerHandler {
	return &CustomerHandler{customerRepo: customerRepo}
}

var ProviderSet = wire.NewSet(NewCustomerHandler)
