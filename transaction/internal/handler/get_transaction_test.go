package handler

import (
	"context"
	"errors"
	"testing"

	"transaction/internal/data"
	"transaction/internal/model"

	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc/codes"
)

func TestGetTransaction_Success(t *testing.T) {
	tx := sampleCompletedTx()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, id int64) (*model.Transaction, error) {
			if id == tx.ID {
				return tx, nil
			}
			return nil, data.ErrNotFound
		},
	}
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).GetTransaction(context.Background(), &transactionv1.GetTransactionRequest{Id: tx.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.Id != tx.ID {
		t.Errorf("expected id %d, got %d", tx.ID, resp.Transaction.Id)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Transaction, error) {
			return nil, data.ErrNotFound
		},
	}
	_, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).GetTransaction(context.Background(), &transactionv1.GetTransactionRequest{Id: 999})
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestGetTransaction_RepoError(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*model.Transaction, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).GetTransaction(context.Background(), &transactionv1.GetTransactionRequest{Id: 1})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
