package handler

import (
	"context"
	"time"

	"transaction/internal/data"
	"transaction/internal/model"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ── mock repo ─────────────────────────────────────────────────────────────────

type mockRepo struct {
	createFn                func(ctx context.Context, tx *model.Transaction) error
	getByIDFn               func(ctx context.Context, id int64) (*model.Transaction, error)
	getByIdempotencyKeyFn   func(ctx context.Context, key string) (*model.Transaction, error)
	markCompletedFn         func(ctx context.Context, id int64, ledgerTxID int64) error
	markFailedFn            func(ctx context.Context, id int64, reason string) error
	listByAccountFn         func(ctx context.Context, accountID int64, p data.ListParams) ([]*model.Transaction, int64, error)
}

func (m *mockRepo) Create(ctx context.Context, tx *model.Transaction) error {
	return m.createFn(ctx, tx)
}
func (m *mockRepo) GetByID(ctx context.Context, id int64) (*model.Transaction, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	return m.getByIdempotencyKeyFn(ctx, key)
}
func (m *mockRepo) MarkCompleted(ctx context.Context, id int64, ledgerTxID int64) error {
	return m.markCompletedFn(ctx, id, ledgerTxID)
}
func (m *mockRepo) MarkFailed(ctx context.Context, id int64, reason string) error {
	if m.markFailedFn != nil {
		return m.markFailedFn(ctx, id, reason)
	}
	return nil
}
func (m *mockRepo) ListByAccount(ctx context.Context, accountID int64, p data.ListParams) ([]*model.Transaction, int64, error) {
	return m.listByAccountFn(ctx, accountID, p)
}

// ── mock account client ───────────────────────────────────────────────────────

type mockAccountClient struct {
	getAccountFn func(ctx context.Context, in *accountv1.GetAccountRequest, opts ...grpc.CallOption) (*accountv1.GetAccountResponse, error)
}

func (m *mockAccountClient) GetAccount(ctx context.Context, in *accountv1.GetAccountRequest, opts ...grpc.CallOption) (*accountv1.GetAccountResponse, error) {
	return m.getAccountFn(ctx, in, opts...)
}

// ── mock ledger client ────────────────────────────────────────────────────────

type mockLedgerClient struct {
	postTransactionFn func(ctx context.Context, in *ledgerv1.PostTransactionRequest, opts ...grpc.CallOption) (*ledgerv1.PostTransactionResponse, error)
}

func (m *mockLedgerClient) PostTransaction(ctx context.Context, in *ledgerv1.PostTransactionRequest, opts ...grpc.CallOption) (*ledgerv1.PostTransactionResponse, error) {
	return m.postTransactionFn(ctx, in, opts...)
}

// ── helpers ───────────────────────────────────────────────────────────────────

const settlementID = int64(1)

func newHandler(repo transactionRepo, accounts data.AccountClient, ledger data.LedgerClient) *TransactionHandler {
	return &TransactionHandler{repo: repo, accounts: accounts, ledger: ledger, settlementAcctID: settlementID}
}

func grpcCode(err error) codes.Code {
	if s, ok := status.FromError(err); ok {
		return s.Code()
	}
	return codes.Unknown
}

// noIdempotencyHit simulates no prior transaction with this key.
func noIdempotencyHit(_ context.Context, _ string) (*model.Transaction, error) {
	return nil, data.ErrNotFound
}

// activeAccountClient returns ACTIVE for any account lookup.
func activeAccountClient() *mockAccountClient {
	return &mockAccountClient{
		getAccountFn: func(_ context.Context, in *accountv1.GetAccountRequest, _ ...grpc.CallOption) (*accountv1.GetAccountResponse, error) {
			return &accountv1.GetAccountResponse{
				Account: &accountv1.Account{
					Id:     in.Id,
					Status: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
				},
			}, nil
		},
	}
}

// successLedgerClient returns a successful PostTransaction with the given ledger tx ID.
func successLedgerClient(ledgerTxID int64) *mockLedgerClient {
	return &mockLedgerClient{
		postTransactionFn: func(_ context.Context, _ *ledgerv1.PostTransactionRequest, _ ...grpc.CallOption) (*ledgerv1.PostTransactionResponse, error) {
			return &ledgerv1.PostTransactionResponse{
				Transaction: &ledgerv1.LedgerTransaction{Id: ledgerTxID},
			}, nil
		},
	}
}

// successRepo returns a repo that assigns the given ID on Create and succeeds on all updates.
func successRepo(assignID int64) *mockRepo {
	return &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		createFn: func(_ context.Context, tx *model.Transaction) error {
			tx.ID = assignID
			return nil
		},
		markCompletedFn: func(_ context.Context, _ int64, _ int64) error { return nil },
	}
}

func sampleCompletedTx() *model.Transaction {
	fromID := int64(10)
	toID := int64(20)
	ledgerID := int64(99)
	state := model.TxStateCompleted
	now := time.Now()
	return &model.Transaction{
		ID:                  1,
		Kind:                model.TxKindTransfer,
		State:               state,
		FromAccountID:       &fromID,
		ToAccountID:         &toID,
		AmountMinor:         1000,
		Currency:            "INR",
		IdempotencyKey:      "idem-001",
		LedgerTransactionID: &ledgerID,
		CreatedAt:           now,
		CompletedAt:         &now,
	}
}

func sampleTransferRequest() *transactionv1.TransferRequest {
	return &transactionv1.TransferRequest{
		FromAccountId:  10,
		ToAccountId:    20,
		AmountMinor:    1000,
		Currency:       "INR",
		IdempotencyKey: "idem-001",
		Description:    "test transfer",
	}
}
