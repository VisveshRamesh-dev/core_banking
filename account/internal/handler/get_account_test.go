package handler

import (
	"context"
	"errors"
	"testing"

	"account/internal/data"
	"account/internal/model"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"google.golang.org/grpc/codes"
)

func TestGetAccount_Success(t *testing.T) {
	acc := sampleAccount()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, id int64) (*model.Account, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, data.ErrNotFound
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).GetAccount(context.Background(), &accountv1.GetAccountRequest{Id: acc.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Account.Id != acc.ID {
		t.Errorf("expected id %d, got %d", acc.ID, resp.Account.Id)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) {
			return nil, data.ErrNotFound
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).GetAccount(context.Background(), &accountv1.GetAccountRequest{Id: 999})
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestGetAccount_RepoError(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Account, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).GetAccount(context.Background(), &accountv1.GetAccountRequest{Id: 1})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
