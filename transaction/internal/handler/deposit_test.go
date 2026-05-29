package handler

import (
	"context"
	"errors"
	"testing"

	"transaction/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestDeposit_Success(t *testing.T) {
	repo := successRepo(2)
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(100)).Deposit(context.Background(), &transactionv1.DepositRequest{
		ToAccountId:     20,
		AmountMinor:     5000,
		Currency:        "INR",
		IdempotencyKey:  "dep-001",
		SourceReference: "UTR123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.State != commonv1.TransactionState_TRANSACTION_STATE_COMPLETED {
		t.Errorf("expected COMPLETED, got %s", resp.Transaction.State)
	}
}

func TestDeposit_IdempotencyHit(t *testing.T) {
	existing := sampleCompletedTx()
	existing.Kind = model.TxKindDeposit
	repo := &mockRepo{
		getByIdempotencyKeyFn: func(_ context.Context, _ string) (*model.Transaction, error) {
			return existing, nil
		},
	}
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).Deposit(context.Background(), &transactionv1.DepositRequest{
		ToAccountId: 20, AmountMinor: 5000, Currency: "INR", IdempotencyKey: "dep-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.Id != existing.ID {
		t.Errorf("expected existing tx id %d, got %d", existing.ID, resp.Transaction.Id)
	}
}

func TestDeposit_MissingIdempotencyKey(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).Deposit(context.Background(), &transactionv1.DepositRequest{
		ToAccountId: 20, AmountMinor: 5000, Currency: "INR",
	})
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestDeposit_ZeroAmount(t *testing.T) {
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).Deposit(context.Background(), &transactionv1.DepositRequest{
		ToAccountId: 20, AmountMinor: 0, Currency: "INR", IdempotencyKey: "dep-002",
	})
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestDeposit_LedgerUsesSettlementAccount(t *testing.T) {
	var capturedEntries []*ledgerv1.Entry
	ledger := &mockLedgerClient{
		postTransactionFn: func(_ context.Context, in *ledgerv1.PostTransactionRequest, _ ...grpc.CallOption) (*ledgerv1.PostTransactionResponse, error) {
			capturedEntries = in.Entries
			return &ledgerv1.PostTransactionResponse{Transaction: &ledgerv1.LedgerTransaction{Id: 1}}, nil
		},
	}
	repo := successRepo(1)
	_, _ = newHandler(repo, activeAccountClient(), ledger).Deposit(context.Background(), &transactionv1.DepositRequest{
		ToAccountId: 20, AmountMinor: 5000, Currency: "INR", IdempotencyKey: "dep-003",
	})
	if len(capturedEntries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(capturedEntries))
	}
	// First entry should debit the settlement account.
	if capturedEntries[0].AccountId != settlementID {
		t.Errorf("expected settlement account %d debited, got %d", settlementID, capturedEntries[0].AccountId)
	}
	if capturedEntries[0].Direction != commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT {
		t.Error("expected settlement entry to be DEBIT")
	}
	// Second entry should credit the destination account.
	if capturedEntries[1].AccountId != 20 {
		t.Errorf("expected to_account 20 credited, got %d", capturedEntries[1].AccountId)
	}
	if capturedEntries[1].Direction != commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT {
		t.Error("expected to_account entry to be CREDIT")
	}
}

func TestDeposit_LedgerError_MarksFailed(t *testing.T) {
	var markedFailed bool
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		createFn: func(_ context.Context, tx *model.Transaction) error {
			tx.ID = 1
			return nil
		},
		markFailedFn: func(_ context.Context, _ int64, _ string) error {
			markedFailed = true
			return nil
		},
	}
	errLedger := &mockLedgerClient{
		postTransactionFn: func(_ context.Context, _ *ledgerv1.PostTransactionRequest, _ ...grpc.CallOption) (*ledgerv1.PostTransactionResponse, error) {
			return nil, errors.New("ledger down")
		},
	}
	_, err := newHandler(repo, activeAccountClient(), errLedger).Deposit(context.Background(), &transactionv1.DepositRequest{
		ToAccountId: 20, AmountMinor: 5000, Currency: "INR", IdempotencyKey: "dep-004",
	})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
	if !markedFailed {
		t.Error("expected MarkFailed to be called")
	}
}
