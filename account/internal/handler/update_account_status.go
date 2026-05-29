package handler

import (
	"context"
	"errors"

	"account/internal/data"
	"account/internal/mapper"
	"account/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// validTransitions maps current status → set of allowed next statuses.
// CloseAccount is handled separately (requires zero-balance check).
var validTransitions = map[int16]map[int16]bool{
	model.AccountStatusPending: {model.AccountStatusActive: true},
	model.AccountStatusActive:  {model.AccountStatusFrozen: true},
	model.AccountStatusFrozen:  {model.AccountStatusActive: true},
}

func (h *AccountHandler) UpdateAccountStatus(ctx context.Context, req *accountv1.UpdateAccountStatusRequest) (*accountv1.UpdateAccountStatusResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.NewStatus == commonv1.AccountStatus_ACCOUNT_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "new_status is required")
	}
	newStatus := int16(req.NewStatus)
	if newStatus == model.AccountStatusClosed {
		return nil, status.Error(codes.InvalidArgument, "use CloseAccount to close an account")
	}

	account, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "account %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if allowed := validTransitions[account.Status]; !allowed[newStatus] {
		return nil, status.Errorf(codes.FailedPrecondition,
			"cannot transition account from %s to %s",
			commonv1.AccountStatus(account.Status), req.NewStatus)
	}

	updated, err := h.repo.UpdateStatus(ctx, req.Id, newStatus)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountv1.UpdateAccountStatusResponse{Account: mapper.AccountToProto(updated)}, nil
}
