package handler

import (
	"context"
	"errors"
	"testing"

	"ledger/internal/data"

	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
)

func TestGetTransaction_Success(t *testing.T) {
	rec := samplePostedRecord()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, id int64) (*data.TxRecord, error) {
			if id == rec.Tx.ID {
				return rec, nil
			}
			return nil, data.ErrNotFound
		},
	}
	resp, err := newHandler(repo).GetTransaction(context.Background(), &v1.GetTransactionRequest{TransactionId: rec.Tx.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.Id != rec.Tx.ID {
		t.Errorf("expected tx id %d, got %d", rec.Tx.ID, resp.Transaction.Id)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.TxRecord, error) {
			return nil, data.ErrNotFound
		},
	}
	_, err := newHandler(repo).GetTransaction(context.Background(), &v1.GetTransactionRequest{TransactionId: 999})
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestGetTransaction_RepoError(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.TxRecord, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo).GetTransaction(context.Background(), &v1.GetTransactionRequest{TransactionId: 1})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
