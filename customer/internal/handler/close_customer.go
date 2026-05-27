package handler

import (
	"context"
	"errors"

	"customer/internal/data"
	"customer/internal/mapper"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CustomerHandler) CloseCustomer(ctx context.Context, req *v1.CloseCustomerRequest) (*v1.CloseCustomerResponse, error) {
	rec, err := h.customerRepo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "customer %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	current := commonv1.KYCStatus(rec.Customer.KycStatus)
	if current == commonv1.KYCStatus_KYC_STATUS_CLOSED {
		return nil, status.Error(codes.FailedPrecondition, "customer is already closed")
	}
	if !isAllowed(current, commonv1.KYCStatus_KYC_STATUS_CLOSED) {
		return nil, status.Errorf(codes.FailedPrecondition,
			"customer in status %s cannot be closed", current.String())
	}

	rec, err = h.customerRepo.UpdateKYCStatus(ctx, req.Id, int16(commonv1.KYCStatus_KYC_STATUS_CLOSED))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.CustomerToProto(
		rec.Customer, rec.Phones, rec.Addresses,
		rec.Individual, rec.Business,
		rec.BizPhones, rec.BizAddrs, rec.Proprietor, rec.PropPhones,
	)
	return &v1.CloseCustomerResponse{Customer: proto}, nil
}
