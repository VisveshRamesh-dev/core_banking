package handler

import (
	"context"
	"errors"

	"customer/internal/data"
	"customer/internal/mapper"

	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CustomerHandler) GetCustomer(ctx context.Context, req *v1.GetCustomerRequest) (*v1.GetCustomerResponse, error) {
	rec, err := h.customerRepo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "customer %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.CustomerToProto(
		rec.Customer,
		rec.Individual, rec.Business,
		rec.Phones, rec.Addresses,
		rec.BizPhones, rec.BizAddrs,
		rec.PropPhones,
	)
	return &v1.GetCustomerResponse{Customer: proto}, nil
}
