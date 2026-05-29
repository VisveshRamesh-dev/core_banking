package handler

import (
	"context"
	"errors"
	"testing"

	"ledger/internal/data"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
)

func TestPostTransaction_Success(t *testing.T) {
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		postFn: func(_ context.Context, rec *data.TxRecord) error {
			rec.Tx.ID = 1
			for i := range rec.Entries {
				rec.Entries[i].ID = int64(i + 1)
				rec.Entries[i].TransactionID = 1
			}
			return nil
		},
	}
	resp, err := newHandler(repo).PostTransaction(context.Background(), samplePostRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction == nil {
		t.Fatal("expected transaction in response")
	}
	if len(resp.Transaction.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Transaction.Entries))
	}
}

func TestPostTransaction_IdempotencyHit(t *testing.T) {
	existing := samplePostedRecord()
	repo := &mockRepo{
		getByIdempotencyKeyFn: func(_ context.Context, _ string) (*data.TxRecord, error) {
			return existing, nil
		},
	}
	resp, err := newHandler(repo).PostTransaction(context.Background(), samplePostRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.Id != existing.Tx.ID {
		t.Errorf("expected existing tx id %d, got %d", existing.Tx.ID, resp.Transaction.Id)
	}
}

func TestPostTransaction_MissingIdempotencyKey(t *testing.T) {
	req := samplePostRequest()
	req.IdempotencyKey = ""
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_TooFewEntries(t *testing.T) {
	req := samplePostRequest()
	req.Entries = req.Entries[:1]
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_NegativeAmount(t *testing.T) {
	req := samplePostRequest()
	req.Entries[0].AmountMinor = -100
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_ZeroAmount(t *testing.T) {
	req := samplePostRequest()
	req.Entries[0].AmountMinor = 0
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_MissingAccountID(t *testing.T) {
	req := samplePostRequest()
	req.Entries[0].AccountId = 0
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_MixedCurrencies(t *testing.T) {
	req := &v1.PostTransactionRequest{
		IdempotencyKey: "idem-mix",
		Entries: []*v1.Entry{
			{AccountId: 10, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT, AmountMinor: 1000, Currency: "INR"},
			{AccountId: 20, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT, AmountMinor: 1000, Currency: "USD"},
		},
	}
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_DoesNotBalance(t *testing.T) {
	req := &v1.PostTransactionRequest{
		IdempotencyKey: "idem-unbal",
		Entries: []*v1.Entry{
			{AccountId: 10, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT, AmountMinor: 1000, Currency: "INR"},
			{AccountId: 20, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT, AmountMinor: 900, Currency: "INR"},
		},
	}
	_, err := newHandler(&mockRepo{}).PostTransaction(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestPostTransaction_RepoError(t *testing.T) {
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		postFn: func(_ context.Context, _ *data.TxRecord) error {
			return errors.New("db down")
		},
	}
	_, err := newHandler(repo).PostTransaction(context.Background(), samplePostRequest())
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}

func TestPostTransaction_IdempotencyLookupError(t *testing.T) {
	repo := &mockRepo{
		getByIdempotencyKeyFn: func(_ context.Context, _ string) (*data.TxRecord, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo).PostTransaction(context.Background(), samplePostRequest())
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
