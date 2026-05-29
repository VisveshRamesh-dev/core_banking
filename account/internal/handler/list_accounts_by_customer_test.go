package handler

import (
	"context"
	"errors"
	"testing"

	"account/internal/data"
	"account/internal/model"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	"google.golang.org/grpc/codes"
)

func TestListAccountsByCustomer_Success(t *testing.T) {
	accs := []*model.Account{sampleAccount()}
	repo := &mockRepo{
		listByCustomerFn: func(_ context.Context, customerID int64, _ data.ListParams) ([]*model.Account, int64, error) {
			if customerID == 100 {
				return accs, 1, nil
			}
			return nil, 0, nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).ListAccountsByCustomer(
		context.Background(),
		&accountv1.ListAccountsByCustomerRequest{
			CustomerId: 100,
			Page:       &commonv1.PageRequest{PageSize: 10},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(resp.Accounts))
	}
	if resp.Page.TotalSize != 1 {
		t.Errorf("expected total_size 1, got %d", resp.Page.TotalSize)
	}
}

func TestListAccountsByCustomer_MissingCustomerID(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).ListAccountsByCustomer(
		context.Background(),
		&accountv1.ListAccountsByCustomerRequest{CustomerId: 0},
	)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestListAccountsByCustomer_RepoError(t *testing.T) {
	repo := &mockRepo{
		listByCustomerFn: func(_ context.Context, _ int64, _ data.ListParams) ([]*model.Account, int64, error) {
			return nil, 0, errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).ListAccountsByCustomer(
		context.Background(),
		&accountv1.ListAccountsByCustomerRequest{CustomerId: 100},
	)
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}

func TestListAccountsByCustomer_NextPageToken(t *testing.T) {
	all := []*model.Account{sampleAccount(), sampleAccount(), sampleAccount()}
	all[1].ID = 2
	all[2].ID = 3
	repo := &mockRepo{
		listByCustomerFn: func(_ context.Context, _ int64, p data.ListParams) ([]*model.Account, int64, error) {
			end := p.Offset + p.Limit
			if end > len(all) {
				end = len(all)
			}
			return all[p.Offset:end], int64(len(all)), nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).ListAccountsByCustomer(
		context.Background(),
		&accountv1.ListAccountsByCustomerRequest{
			CustomerId: 100,
			Page:       &commonv1.PageRequest{PageSize: 2},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Accounts) != 2 {
		t.Errorf("expected 2 accounts on first page, got %d", len(resp.Accounts))
	}
	if resp.Page.NextPageToken == "" {
		t.Error("expected next_page_token to be set")
	}
}
