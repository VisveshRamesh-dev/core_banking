package handler

import (
	"context"
	"errors"

	"account/internal/data"
	"account/internal/mapper"
	"account/internal/model"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AccountHandler) CloseAccount(ctx context.Context, req *accountv1.CloseAccountRequest) (*accountv1.CloseAccountResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	account, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "account %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if account.Status == model.AccountStatusClosed {
		return nil, status.Error(codes.FailedPrecondition, "account is already closed")
	}
	if account.Status != model.AccountStatusActive {
		return nil, status.Errorf(codes.FailedPrecondition,
			"only ACTIVE accounts can be closed (current status: %s)",
			commonv1AccountStatus(account.Status))
	}

	// Authoritative balance check via the ledger service.
	balResp, err := h.ledger.GetBalance(ctx, &ledgerv1.GetBalanceRequest{AccountId: req.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if balResp.Balance != nil && balResp.Balance.Balance != nil && balResp.Balance.Balance.AmountMinor != 0 {
		return nil, status.Errorf(codes.FailedPrecondition,
			"account balance must be zero before closing (current balance: %d %s)",
			balResp.Balance.Balance.AmountMinor, balResp.Balance.Balance.Currency)
	}

	updated, err := h.repo.UpdateStatus(ctx, req.Id, model.AccountStatusClosed)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountv1.CloseAccountResponse{Account: mapper.AccountToProto(updated)}, nil
}

// commonv1AccountStatus returns a human-readable status name for error messages.
func commonv1AccountStatus(s int16) string {
	switch s {
	case model.AccountStatusPending:
		return "PENDING"
	case model.AccountStatusActive:
		return "ACTIVE"
	case model.AccountStatusFrozen:
		return "FROZEN"
	case model.AccountStatusClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}
