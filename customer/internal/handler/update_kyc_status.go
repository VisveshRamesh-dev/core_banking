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

// validTransitions defines the allowed KYC state-machine moves.
//
//	PENDING → VERIFIED → ACTIVE → SUSPENDED → ACTIVE
//	                        └──→ CLOSED (terminal)
var validTransitions = map[commonv1.KYCStatus][]commonv1.KYCStatus{
	commonv1.KYCStatus_KYC_STATUS_PENDING:   {commonv1.KYCStatus_KYC_STATUS_VERIFIED},
	commonv1.KYCStatus_KYC_STATUS_VERIFIED:  {commonv1.KYCStatus_KYC_STATUS_ACTIVE},
	commonv1.KYCStatus_KYC_STATUS_ACTIVE:    {commonv1.KYCStatus_KYC_STATUS_SUSPENDED, commonv1.KYCStatus_KYC_STATUS_CLOSED},
	commonv1.KYCStatus_KYC_STATUS_SUSPENDED: {commonv1.KYCStatus_KYC_STATUS_ACTIVE},
}

func (h *CustomerHandler) UpdateKYCStatus(ctx context.Context, req *v1.UpdateKYCStatusRequest) (*v1.UpdateKYCStatusResponse, error) {
	rec, err := h.customerRepo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "customer %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	current := commonv1.KYCStatus(rec.Customer.KycStatus)
	if !isAllowed(current, req.NewStatus) {
		return nil, status.Errorf(codes.FailedPrecondition,
			"transition from %s to %s is not allowed",
			current.String(), req.NewStatus.String())
	}

	rec, err = h.customerRepo.UpdateKYCStatus(ctx, req.Id, int16(req.NewStatus))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.CustomerToProto(
		rec.Customer,
		rec.Individual, rec.Business,
		rec.Phones, rec.Addresses,
		rec.BizPhones, rec.BizAddrs,
		rec.PropPhones,
	)
	return &v1.UpdateKYCStatusResponse{Customer: proto}, nil
}

func isAllowed(from, to commonv1.KYCStatus) bool {
	for _, allowed := range validTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
