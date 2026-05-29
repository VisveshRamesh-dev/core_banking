package service

import (
	"context"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"github.com/google/wire"

	"account/internal/handler"
)

var ProviderSet = wire.NewSet(NewAccountService)

type AccountService struct {
	accountv1.UnimplementedAccountServiceServer
	handler *handler.AccountHandler
}

func NewAccountService(h *handler.AccountHandler) *AccountService {
	return &AccountService{handler: h}
}

func (s *AccountService) OpenAccount(ctx context.Context, req *accountv1.OpenAccountRequest) (*accountv1.OpenAccountResponse, error) {
	return s.handler.OpenAccount(ctx, req)
}

func (s *AccountService) GetAccount(ctx context.Context, req *accountv1.GetAccountRequest) (*accountv1.GetAccountResponse, error) {
	return s.handler.GetAccount(ctx, req)
}

func (s *AccountService) ListAccountsByCustomer(ctx context.Context, req *accountv1.ListAccountsByCustomerRequest) (*accountv1.ListAccountsByCustomerResponse, error) {
	return s.handler.ListAccountsByCustomer(ctx, req)
}

func (s *AccountService) UpdateAccountStatus(ctx context.Context, req *accountv1.UpdateAccountStatusRequest) (*accountv1.UpdateAccountStatusResponse, error) {
	return s.handler.UpdateAccountStatus(ctx, req)
}

func (s *AccountService) CloseAccount(ctx context.Context, req *accountv1.CloseAccountRequest) (*accountv1.CloseAccountResponse, error) {
	return s.handler.CloseAccount(ctx, req)
}
