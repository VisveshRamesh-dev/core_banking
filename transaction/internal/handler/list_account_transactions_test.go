package handler

import (
	"context"
	"errors"
	"testing"

	"transaction/internal/data"
	"transaction/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc/codes"
)

func TestListAccountTransactions_Success(t *testing.T) {
	txs := []*model.Transaction{sampleCompletedTx()}
	repo := &mockRepo{
		listByAccountFn: func(_ context.Context, accountID int64, _ data.ListParams) ([]*model.Transaction, int64, error) {
			if accountID == 10 {
				return txs, 1, nil
			}
			return nil, 0, nil
		},
	}
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).ListAccountTransactions(
		context.Background(),
		&transactionv1.ListAccountTransactionsRequest{
			AccountId: 10,
			Page:      &commonv1.PageRequest{PageSize: 10},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(resp.Transactions))
	}
	if resp.Page.TotalSize != 1 {
		t.Errorf("expected total_size 1, got %d", resp.Page.TotalSize)
	}
}

func TestListAccountTransactions_MissingAccountID(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).ListAccountTransactions(
		context.Background(),
		&transactionv1.ListAccountTransactionsRequest{AccountId: 0},
	)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestListAccountTransactions_RepoError(t *testing.T) {
	repo := &mockRepo{
		listByAccountFn: func(_ context.Context, _ int64, _ data.ListParams) ([]*model.Transaction, int64, error) {
			return nil, 0, errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).ListAccountTransactions(
		context.Background(),
		&transactionv1.ListAccountTransactionsRequest{AccountId: 10},
	)
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}

func TestListAccountTransactions_Pagination(t *testing.T) {
	all := []*model.Transaction{sampleCompletedTx(), sampleCompletedTx(), sampleCompletedTx()}
	all[1].ID = 2
	all[2].ID = 3
	repo := &mockRepo{
		listByAccountFn: func(_ context.Context, _ int64, p data.ListParams) ([]*model.Transaction, int64, error) {
			end := p.Offset + p.Limit
			if end > len(all) {
				end = len(all)
			}
			return all[p.Offset:end], int64(len(all)), nil
		},
	}
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).ListAccountTransactions(
		context.Background(),
		&transactionv1.ListAccountTransactionsRequest{
			AccountId: 10,
			Page:      &commonv1.PageRequest{PageSize: 2},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 2 {
		t.Errorf("expected 2 transactions on first page, got %d", len(resp.Transactions))
	}
	if resp.Page.NextPageToken == "" {
		t.Error("expected next_page_token to be set")
	}
}
