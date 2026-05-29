package handler

import (
	"context"
	"time"

	"ledger/internal/data"
	"ledger/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ── mock repo ────────────────────────────────────────────────────────────────

type mockRepo struct {
	postFn                func(ctx context.Context, rec *data.TxRecord) error
	getByIDFn             func(ctx context.Context, id int64) (*data.TxRecord, error)
	getByIdempotencyKeyFn func(ctx context.Context, key string) (*data.TxRecord, error)
	getBalanceFn          func(ctx context.Context, accountID int64) (*data.BalanceResult, error)
}

func (m *mockRepo) Post(ctx context.Context, rec *data.TxRecord) error {
	return m.postFn(ctx, rec)
}
func (m *mockRepo) GetByID(ctx context.Context, id int64) (*data.TxRecord, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) GetByIdempotencyKey(ctx context.Context, key string) (*data.TxRecord, error) {
	return m.getByIdempotencyKeyFn(ctx, key)
}
func (m *mockRepo) GetBalance(ctx context.Context, accountID int64) (*data.BalanceResult, error) {
	return m.getBalanceFn(ctx, accountID)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func newHandler(repo ledgerRepo) *LedgerHandler {
	return &LedgerHandler{repo: repo}
}

func grpcCode(err error) codes.Code {
	if s, ok := status.FromError(err); ok {
		return s.Code()
	}
	return codes.Unknown
}

// noIdempotencyHit returns ErrNotFound for the idempotency key lookup (common case: first call).
func noIdempotencyHit(_ context.Context, _ string) (*data.TxRecord, error) {
	return nil, data.ErrNotFound
}

// samplePostedRecord returns a minimal TxRecord that represents a posted transaction.
func samplePostedRecord() *data.TxRecord {
	desc := "payment"
	origID := int64(0)
	_ = origID
	return &data.TxRecord{
		Tx: model.LedgerTransaction{
			ID:             1,
			Status:         model.TxStatusPosted,
			Description:    &desc,
			IdempotencyKey: "idem-001",
			PostedAt:       time.Now(),
		},
		Entries: []model.LedgerEntry{
			{ID: 1, TransactionID: 1, AccountID: 10, Direction: model.DirectionDebit, AmountMinor: 1000, Currency: "INR"},
			{ID: 2, TransactionID: 1, AccountID: 20, Direction: model.DirectionCredit, AmountMinor: 1000, Currency: "INR"},
		},
	}
}

// samplePostRequest returns a valid balanced PostTransactionRequest.
func samplePostRequest() *v1.PostTransactionRequest {
	return &v1.PostTransactionRequest{
		IdempotencyKey: "idem-001",
		Description:    "payment",
		Entries: []*v1.Entry{
			{AccountId: 10, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT, AmountMinor: 1000, Currency: "INR"},
			{AccountId: 20, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT, AmountMinor: 1000, Currency: "INR"},
		},
	}
}
