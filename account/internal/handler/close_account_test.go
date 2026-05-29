package handler

import (
	"context"
	"errors"
	"testing"

	"account/internal/data"
	"account/internal/model"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestCloseAccount_Success(t *testing.T) {
	acc := sampleAccount() // ACTIVE
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
		updateStatusFn: func(_ context.Context, _ int64, s int16) (*model.Account, error) {
			acc.Status = s
			return acc, nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 1, Reason: "customer request"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Account == nil {
		t.Fatal("expected account in response")
	}
}

func TestCloseAccount_MissingID(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 0},
	)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestCloseAccount_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return nil, data.ErrNotFound },
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 999},
	)
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestCloseAccount_AlreadyClosed(t *testing.T) {
	acc := sampleAccount()
	acc.Status = model.AccountStatusClosed
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 1},
	)
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestCloseAccount_NotActive(t *testing.T) {
	acc := sampleAccount()
	acc.Status = model.AccountStatusFrozen
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 1},
	)
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestCloseAccount_NonZeroBalance(t *testing.T) {
	acc := sampleAccount() // ACTIVE
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
	}
	_, err := newHandler(repo, activeCustomerClient(), nonZeroBalanceLedgerClient(5000)).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 1},
	)
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestCloseAccount_LedgerError(t *testing.T) {
	acc := sampleAccount()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
	}
	errLedger := &mockLedgerClient{
		getBalanceFn: func(_ context.Context, _ *ledgerv1.GetBalanceRequest, _ ...grpc.CallOption) (*ledgerv1.GetBalanceResponse, error) {
			return nil, errors.New("ledger down")
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), errLedger).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 1},
	)
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}

func TestCloseAccount_RepoUpdateError(t *testing.T) {
	acc := sampleAccount()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) { return acc, nil },
		updateStatusFn: func(_ context.Context, _ int64, _ int16) (*model.Account, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).CloseAccount(
		context.Background(),
		&accountv1.CloseAccountRequest{Id: 1},
	)
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
