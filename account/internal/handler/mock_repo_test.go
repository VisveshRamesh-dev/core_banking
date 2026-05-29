package handler

import (
	"context"
	"time"

	"account/internal/data"
	"account/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	customerv1 "github.com/visvesh-ramesh/corebank/v1/customer"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ── mock repo ────────────────────────────────────────────────────────────────

type mockRepo struct {
	createFn         func(ctx context.Context, a *model.Account) error
	getByIDFn        func(ctx context.Context, id int64) (*model.Account, error)
	updateStatusFn   func(ctx context.Context, id int64, newStatus int16) (*model.Account, error)
	listByCustomerFn func(ctx context.Context, customerID int64, p data.ListParams) ([]*model.Account, int64, error)
}

func (m *mockRepo) Create(ctx context.Context, a *model.Account) error {
	return m.createFn(ctx, a)
}
func (m *mockRepo) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) UpdateStatus(ctx context.Context, id int64, newStatus int16) (*model.Account, error) {
	return m.updateStatusFn(ctx, id, newStatus)
}
func (m *mockRepo) ListByCustomer(ctx context.Context, customerID int64, p data.ListParams) ([]*model.Account, int64, error) {
	return m.listByCustomerFn(ctx, customerID, p)
}

// ── mock customer client ──────────────────────────────────────────────────────

type mockCustomerClient struct {
	getCustomerFn func(ctx context.Context, in *customerv1.GetCustomerRequest, opts ...grpc.CallOption) (*customerv1.GetCustomerResponse, error)
}

func (m *mockCustomerClient) GetCustomer(ctx context.Context, in *customerv1.GetCustomerRequest, opts ...grpc.CallOption) (*customerv1.GetCustomerResponse, error) {
	return m.getCustomerFn(ctx, in, opts...)
}

// ── mock ledger client ────────────────────────────────────────────────────────

type mockLedgerClient struct {
	getBalanceFn func(ctx context.Context, in *ledgerv1.GetBalanceRequest, opts ...grpc.CallOption) (*ledgerv1.GetBalanceResponse, error)
}

func (m *mockLedgerClient) GetBalance(ctx context.Context, in *ledgerv1.GetBalanceRequest, opts ...grpc.CallOption) (*ledgerv1.GetBalanceResponse, error) {
	return m.getBalanceFn(ctx, in, opts...)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func newHandler(repo accountRepo, customers data.CustomerClient, ledger data.LedgerClient) *AccountHandler {
	return &AccountHandler{repo: repo, customers: customers, ledger: ledger}
}

func grpcCode(err error) codes.Code {
	if s, ok := status.FromError(err); ok {
		return s.Code()
	}
	return codes.Unknown
}

func activeCustomerClient() *mockCustomerClient {
	return &mockCustomerClient{
		getCustomerFn: func(_ context.Context, _ *customerv1.GetCustomerRequest, _ ...grpc.CallOption) (*customerv1.GetCustomerResponse, error) {
			return &customerv1.GetCustomerResponse{
				Customer: &customerv1.Customer{
					Id:        1,
					KycStatus: commonv1.KYCStatus_KYC_STATUS_ACTIVE,
				},
			}, nil
		},
	}
}

func zeroBalanceLedgerClient() *mockLedgerClient {
	return &mockLedgerClient{
		getBalanceFn: func(_ context.Context, _ *ledgerv1.GetBalanceRequest, _ ...grpc.CallOption) (*ledgerv1.GetBalanceResponse, error) {
			return &ledgerv1.GetBalanceResponse{
				Balance: &ledgerv1.Balance{
					Balance: &commonv1.Money{AmountMinor: 0, Currency: "INR"},
				},
			}, nil
		},
	}
}

func nonZeroBalanceLedgerClient(amount int64) *mockLedgerClient {
	return &mockLedgerClient{
		getBalanceFn: func(_ context.Context, _ *ledgerv1.GetBalanceRequest, _ ...grpc.CallOption) (*ledgerv1.GetBalanceResponse, error) {
			return &ledgerv1.GetBalanceResponse{
				Balance: &ledgerv1.Balance{
					Balance: &commonv1.Money{AmountMinor: amount, Currency: "INR"},
				},
			}, nil
		},
	}
}

func sampleAccount() *model.Account {
	return &model.Account{
		ID:         1,
		CustomerID: 100,
		Type:       model.AccountTypeSavings,
		Status:     model.AccountStatusActive,
		Currency:   "INR",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func sampleOpenRequest() *accountv1.OpenAccountRequest {
	return &accountv1.OpenAccountRequest{
		CustomerId: 100,
		Type:       commonv1.AccountType_ACCOUNT_TYPE_SAVINGS,
		Currency:   "INR",
	}
}
