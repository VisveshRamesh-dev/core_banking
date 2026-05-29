package handler

import (
	"context"
	"errors"

	"account/internal/data"
	"account/internal/mapper"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AccountHandler) GetAccount(ctx context.Context, req *accountv1.GetAccountRequest) (*accountv1.GetAccountResponse, error) {
	account, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "account %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountv1.GetAccountResponse{Account: mapper.AccountToProto(account)}, nil
}
