package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"ledger/internal/data"

	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
)

func TestGetBalance_Success(t *testing.T) {
	result := &data.BalanceResult{Balance: 5000, Currency: "INR", AsOf: time.Now()}
	repo := &mockRepo{
		getBalanceFn: func(_ context.Context, accountID int64) (*data.BalanceResult, error) {
			if accountID == 10 {
				return result, nil
			}
			return nil, errors.New("unexpected account")
		},
	}
	resp, err := newHandler(repo).GetBalance(context.Background(), &v1.GetBalanceRequest{AccountId: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Balance == nil {
		t.Fatal("expected balance in response")
	}
	if resp.Balance.Balance.AmountMinor != result.Balance {
		t.Errorf("amount_minor: got %d, want %d", resp.Balance.Balance.AmountMinor, result.Balance)
	}
	if resp.Balance.Balance.Currency != result.Currency {
		t.Errorf("currency: got %q, want %q", resp.Balance.Balance.Currency, result.Currency)
	}
	if resp.Balance.AccountId != 10 {
		t.Errorf("account_id: got %d, want 10", resp.Balance.AccountId)
	}
}

func TestGetBalance_MissingAccountID(t *testing.T) {
	_, err := newHandler(&mockRepo{}).GetBalance(context.Background(), &v1.GetBalanceRequest{AccountId: 0})
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestGetBalance_RepoError(t *testing.T) {
	repo := &mockRepo{
		getBalanceFn: func(_ context.Context, _ int64) (*data.BalanceResult, error) {
			return nil, errors.New("db down")
		},
	}
	_, err := newHandler(repo).GetBalance(context.Background(), &v1.GetBalanceRequest{AccountId: 10})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
