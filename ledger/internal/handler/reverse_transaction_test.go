package handler

import (
	"context"
	"errors"
	"testing"

	"ledger/internal/data"
	"ledger/internal/model"

	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
)

func TestReverseTransaction_Success(t *testing.T) {
	original := samplePostedRecord()
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		getByIDFn: func(_ context.Context, id int64) (*data.TxRecord, error) {
			if id == original.Tx.ID {
				return original, nil
			}
			return nil, data.ErrNotFound
		},
		postFn: func(_ context.Context, rec *data.TxRecord) error {
			rec.Tx.ID = 2
			for i := range rec.Entries {
				rec.Entries[i].ID = int64(i + 10)
				rec.Entries[i].TransactionID = 2
			}
			return nil
		},
	}
	req := &v1.ReverseTransactionRequest{
		TransactionId:  original.Tx.ID,
		IdempotencyKey: "rev-idem-001",
		Reason:         "customer request",
	}
	resp, err := newHandler(repo).ReverseTransaction(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction == nil {
		t.Fatal("expected transaction in response")
	}
	if resp.Transaction.ReversesTransactionId == nil {
		t.Error("expected reverses_transaction_id to be set")
	}
	if *resp.Transaction.ReversesTransactionId != original.Tx.ID {
		t.Errorf("reverses_transaction_id: got %d, want %d", *resp.Transaction.ReversesTransactionId, original.Tx.ID)
	}
	// Entries should be flipped.
	if len(resp.Transaction.Entries) != len(original.Entries) {
		t.Errorf("expected %d entries, got %d", len(original.Entries), len(resp.Transaction.Entries))
	}
}

func TestReverseTransaction_IdempotencyHit(t *testing.T) {
	existing := samplePostedRecord()
	existing.Tx.ID = 2
	repo := &mockRepo{
		getByIdempotencyKeyFn: func(_ context.Context, _ string) (*data.TxRecord, error) {
			return existing, nil
		},
	}
	req := &v1.ReverseTransactionRequest{
		TransactionId:  1,
		IdempotencyKey: "rev-idem-001",
	}
	resp, err := newHandler(repo).ReverseTransaction(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.Id != existing.Tx.ID {
		t.Errorf("expected existing tx id %d, got %d", existing.Tx.ID, resp.Transaction.Id)
	}
}

func TestReverseTransaction_MissingIdempotencyKey(t *testing.T) {
	req := &v1.ReverseTransactionRequest{TransactionId: 1}
	_, err := newHandler(&mockRepo{}).ReverseTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestReverseTransaction_OriginalNotFound(t *testing.T) {
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		getByIDFn: func(_ context.Context, _ int64) (*data.TxRecord, error) {
			return nil, data.ErrNotFound
		},
	}
	req := &v1.ReverseTransactionRequest{TransactionId: 999, IdempotencyKey: "rev-idem-002"}
	_, err := newHandler(repo).ReverseTransaction(context.Background(), req)
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestReverseTransaction_OriginalNotPosted(t *testing.T) {
	rejected := samplePostedRecord()
	rejected.Tx.Status = model.TxStatusRejected
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		getByIDFn: func(_ context.Context, _ int64) (*data.TxRecord, error) {
			return rejected, nil
		},
	}
	req := &v1.ReverseTransactionRequest{TransactionId: 1, IdempotencyKey: "rev-idem-003"}
	_, err := newHandler(repo).ReverseTransaction(context.Background(), req)
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestReverseTransaction_RepoPostError(t *testing.T) {
	original := samplePostedRecord()
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		getByIDFn: func(_ context.Context, _ int64) (*data.TxRecord, error) {
			return original, nil
		},
		postFn: func(_ context.Context, _ *data.TxRecord) error {
			return errors.New("db down")
		},
	}
	req := &v1.ReverseTransactionRequest{TransactionId: 1, IdempotencyKey: "rev-idem-004"}
	_, err := newHandler(repo).ReverseTransaction(context.Background(), req)
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
