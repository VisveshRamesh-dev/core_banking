package handler

import (
	"context"

	"account/internal/mapper"
	"account/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	customerv1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AccountHandler) OpenAccount(ctx context.Context, req *accountv1.OpenAccountRequest) (*accountv1.OpenAccountResponse, error) {
	if req.CustomerId == 0 {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	if req.Type == commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "account type is required")
	}
	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	// Validate that the customer exists and is ACTIVE.
	custResp, err := h.customers.GetCustomer(ctx, &customerv1.GetCustomerRequest{Id: req.CustomerId})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, status.Errorf(codes.FailedPrecondition, "customer %d not found", req.CustomerId)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if custResp.Customer.KycStatus != commonv1.KYCStatus_KYC_STATUS_ACTIVE {
		return nil, status.Errorf(codes.FailedPrecondition,
			"customer %d is not ACTIVE (current kyc_status: %s)",
			req.CustomerId, custResp.Customer.KycStatus)
	}

	overdraft := req.OverdraftLimitMinor
	if req.Type != commonv1.AccountType_ACCOUNT_TYPE_CURRENT {
		overdraft = 0
	}

	account := &model.Account{
		CustomerID:          req.CustomerId,
		Type:                int16(req.Type),
		Status:              model.AccountStatusPending,
		Currency:            req.Currency,
		OverdraftLimitMinor: overdraft,
	}

	if err := h.repo.Create(ctx, account); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountv1.OpenAccountResponse{Account: mapper.AccountToProto(account)}, nil
}
