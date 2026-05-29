package handler

import (
	"context"
	"errors"
	"testing"

	"transaction/internal/model"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

func TestTransfer_Success(t *testing.T) {
	repo := successRepo(1)
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(99)).Transfer(context.Background(), sampleTransferRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.State != commonv1.TransactionState_TRANSACTION_STATE_COMPLETED {
		t.Errorf("expected COMPLETED, got %s", resp.Transaction.State)
	}
	if resp.Transaction.LedgerTransactionId == nil || *resp.Transaction.LedgerTransactionId != 99 {
		t.Error("expected ledger_transaction_id=99")
	}
}

func TestTransfer_IdempotencyHit(t *testing.T) {
	existing := sampleCompletedTx()
	repo := &mockRepo{
		getByIdempotencyKeyFn: func(_ context.Context, _ string) (*model.Transaction, error) {
			return existing, nil
		},
	}
	resp, err := newHandler(repo, activeAccountClient(), successLedgerClient(99)).Transfer(context.Background(), sampleTransferRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Transaction.Id != existing.ID {
		t.Errorf("expected existing tx id %d, got %d", existing.ID, resp.Transaction.Id)
	}
}

func TestTransfer_MissingIdempotencyKey(t *testing.T) {
	req := sampleTransferRequest()
	req.IdempotencyKey = ""
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).Transfer(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestTransfer_SameAccount(t *testing.T) {
	req := sampleTransferRequest()
	req.ToAccountId = req.FromAccountId
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).Transfer(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestTransfer_ZeroAmount(t *testing.T) {
	req := sampleTransferRequest()
	req.AmountMinor = 0
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).Transfer(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestTransfer_MissingCurrency(t *testing.T) {
	req := sampleTransferRequest()
	req.Currency = ""
	_, err := newHandler(&mockRepo{}, activeAccountClient(), successLedgerClient(1)).Transfer(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestTransfer_FromAccountNotActive(t *testing.T) {
	frozen := &mockAccountClient{
		getAccountFn: func(_ context.Context, in *accountv1.GetAccountRequest, _ ...grpc.CallOption) (*accountv1.GetAccountResponse, error) {
			if in.Id == 10 {
				return &accountv1.GetAccountResponse{
					Account: &accountv1.Account{Id: in.Id, Status: commonv1.AccountStatus_ACCOUNT_STATUS_FROZEN},
				}, nil
			}
			return &accountv1.GetAccountResponse{
				Account: &accountv1.Account{Id: in.Id, Status: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
			}, nil
		},
	}
	_, err := newHandler(successRepo(1), frozen, successLedgerClient(1)).Transfer(context.Background(), sampleTransferRequest())
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestTransfer_ToAccountNotFound(t *testing.T) {
	notFound := &mockAccountClient{
		getAccountFn: func(_ context.Context, in *accountv1.GetAccountRequest, _ ...grpc.CallOption) (*accountv1.GetAccountResponse, error) {
			if in.Id == 20 {
				return nil, gstatus.Error(codes.NotFound, "not found")
			}
			return &accountv1.GetAccountResponse{
				Account: &accountv1.Account{Id: in.Id, Status: commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE},
			}, nil
		},
	}
	_, err := newHandler(successRepo(1), notFound, successLedgerClient(1)).Transfer(context.Background(), sampleTransferRequest())
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestTransfer_LedgerError_MarksFailed(t *testing.T) {
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
			return nil, errors.New("ledger unavailable")
		},
	}
	_, err := newHandler(repo, activeAccountClient(), errLedger).Transfer(context.Background(), sampleTransferRequest())
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
	if !markedFailed {
		t.Error("expected MarkFailed to be called")
	}
}

func TestTransfer_RepoCreateError(t *testing.T) {
	repo := &mockRepo{
		getByIdempotencyKeyFn: noIdempotencyHit,
		createFn: func(_ context.Context, _ *model.Transaction) error {
			return errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeAccountClient(), successLedgerClient(1)).Transfer(context.Background(), sampleTransferRequest())
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
