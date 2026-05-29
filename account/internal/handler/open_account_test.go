package handler

import (
	"context"
	"errors"
	"testing"

	"account/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	customerv1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

func TestOpenAccount_Success(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, a *model.Account) error {
			a.ID = 1
			return nil
		},
	}
	resp, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), sampleOpenRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Account == nil {
		t.Fatal("expected account in response")
	}
	if resp.Account.Status != commonv1.AccountStatus_ACCOUNT_STATUS_PENDING {
		t.Errorf("expected PENDING status, got %s", resp.Account.Status)
	}
}

func TestOpenAccount_MissingCustomerID(t *testing.T) {
	req := sampleOpenRequest()
	req.CustomerId = 0
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOpenAccount_MissingType(t *testing.T) {
	req := sampleOpenRequest()
	req.Type = commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOpenAccount_MissingCurrency(t *testing.T) {
	req := sampleOpenRequest()
	req.Currency = ""
	_, err := newHandler(&mockRepo{}, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOpenAccount_CustomerNotFound(t *testing.T) {
	cust := &mockCustomerClient{
		getCustomerFn: func(_ context.Context, _ *customerv1.GetCustomerRequest, _ ...grpc.CallOption) (*customerv1.GetCustomerResponse, error) {
			return nil, gstatus.Error(codes.NotFound, "not found")
		},
	}
	_, err := newHandler(&mockRepo{}, cust, zeroBalanceLedgerClient()).OpenAccount(context.Background(), sampleOpenRequest())
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestOpenAccount_CustomerNotActive(t *testing.T) {
	cust := &mockCustomerClient{
		getCustomerFn: func(_ context.Context, _ *customerv1.GetCustomerRequest, _ ...grpc.CallOption) (*customerv1.GetCustomerResponse, error) {
			return &customerv1.GetCustomerResponse{
				Customer: &customerv1.Customer{Id: 1, KycStatus: commonv1.KYCStatus_KYC_STATUS_PENDING},
			}, nil
		},
	}
	_, err := newHandler(&mockRepo{}, cust, zeroBalanceLedgerClient()).OpenAccount(context.Background(), sampleOpenRequest())
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestOpenAccount_OverdraftStrippedForSavings(t *testing.T) {
	var created model.Account
	repo := &mockRepo{
		createFn: func(_ context.Context, a *model.Account) error {
			created = *a
			return nil
		},
	}
	req := sampleOpenRequest()
	req.OverdraftLimitMinor = 50000
	_, _ = newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), req)
	if created.OverdraftLimitMinor != 0 {
		t.Errorf("overdraft should be stripped for SAVINGS, got %d", created.OverdraftLimitMinor)
	}
}

func TestOpenAccount_OverdraftKeptForCurrent(t *testing.T) {
	var created model.Account
	repo := &mockRepo{
		createFn: func(_ context.Context, a *model.Account) error {
			created = *a
			return nil
		},
	}
	req := &accountv1.OpenAccountRequest{
		CustomerId:          100,
		Type:                commonv1.AccountType_ACCOUNT_TYPE_CURRENT,
		Currency:            "INR",
		OverdraftLimitMinor: 50000,
	}
	_, _ = newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), req)
	if created.OverdraftLimitMinor != 50000 {
		t.Errorf("overdraft should be kept for CURRENT, got %d", created.OverdraftLimitMinor)
	}
}

func TestOpenAccount_RepoError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *model.Account) error {
			return errors.New("db down")
		},
	}
	_, err := newHandler(repo, activeCustomerClient(), zeroBalanceLedgerClient()).OpenAccount(context.Background(), sampleOpenRequest())
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
