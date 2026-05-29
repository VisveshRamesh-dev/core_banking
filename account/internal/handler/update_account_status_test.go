package handler

import (
	"context"
	"errors"
	"testing"

	"account/internal/data"
	"account/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"google.golang.org/grpc/codes"
)

func TestUpdateAccountStatus_PendingToActive(t *testing.T) {
	acc := sampleAccount()
	acc.Status = model.AccountStatusPending
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
		updateStatusFn: func(_ context.Context, _ int64, s int16) (*model.Account, error) {
			acc.Status = s
			return acc, nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 1, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Account.Status != commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE {
		t.Errorf("expected ACTIVE, got %s", resp.Account.Status)
	}
}

func TestUpdateAccountStatus_ActiveToFrozen(t *testing.T) {
	acc := sampleAccount() // already ACTIVE
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
		updateStatusFn: func(_ context.Context, _ int64, s int16) (*model.Account, error) {
			acc.Status = s
			return acc, nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 1, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_FROZEN},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Account.Status != commonv1.AccountStatus_ACCOUNT_STATUS_FROZEN {
		t.Errorf("expected FROZEN, got %s", resp.Account.Status)
	}
}

func TestUpdateAccountStatus_FrozenToActive(t *testing.T) {
	acc := sampleAccount()
	acc.Status = model.AccountStatusFrozen
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
		updateStatusFn: func(_ context.Context, _ int64, s int16) (*model.Account, error) {
			acc.Status = s
			return acc, nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 1, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Account.Status != commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE {
		t.Errorf("expected ACTIVE, got %s", resp.Account.Status)
	}
}

func TestUpdateAccountStatus_InvalidTransition(t *testing.T) {
	acc := sampleAccount() // ACTIVE
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
	}
	// ACTIVE → PENDING is not allowed
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 1, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_PENDING},
	)
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestUpdateAccountStatus_RejectClosedViaUpdate(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 1, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_CLOSED},
	)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestUpdateAccountStatus_MissingID(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 0, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
	)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestUpdateAccountStatus_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return nil, data.ErrNotFound },
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 999, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
	)
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestUpdateAccountStatus_RepoError(t *testing.T) {
	acc := sampleAccount()
	acc.Status = model.AccountStatusPending
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
		updateStatusFn: func(_ context.Context, _ int64, _ int16) (*model.Account, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).UpdateAccountStatus(
		context.Background(),
		&accountv1.UpdateAccountStatusRequest{Id: 1, NewStatus: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
	)
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
