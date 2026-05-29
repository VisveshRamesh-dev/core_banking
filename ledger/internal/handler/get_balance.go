package handler

import (
	"context"

	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *LedgerHandler) GetBalance(ctx context.Context, req *v1.GetBalanceRequest) (*v1.GetBalanceResponse, error) {
	if req.AccountId == 0 {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	result, err := h.repo.GetBalance(ctx, req.AccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.GetBalanceResponse{
		Balance: &v1.Balance{
			AccountId: req.AccountId,
			Balance: &commonv1.Money{
				AmountMinor: result.Balance,
				Currency:    result.Currency,
			},
			AsOf: timestamppb.New(result.AsOf),
		},
	}, nil
}
