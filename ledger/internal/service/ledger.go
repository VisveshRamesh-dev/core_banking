package service

import (
	"context"

	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"github.com/google/wire"

	"ledger/internal/handler"
)

var ProviderSet = wire.NewSet(NewLedgerService)

type LedgerService struct {
	v1.UnimplementedLedgerServiceServer
	handler *handler.LedgerHandler
}

func NewLedgerService(handler *handler.LedgerHandler) *LedgerService {
	return &LedgerService{handler: handler}
}

func (s *LedgerService) PostTransaction(ctx context.Context, req *v1.PostTransactionRequest) (*v1.PostTransactionResponse, error) {
	return s.handler.PostTransaction(ctx, req)
}

func (s *LedgerService) ReverseTransaction(ctx context.Context, req *v1.ReverseTransactionRequest) (*v1.ReverseTransactionResponse, error) {
	return s.handler.ReverseTransaction(ctx, req)
}

func (s *LedgerService) GetTransaction(ctx context.Context, req *v1.GetTransactionRequest) (*v1.GetTransactionResponse, error) {
	return s.handler.GetTransaction(ctx, req)
}

func (s *LedgerService) GetBalance(ctx context.Context, req *v1.GetBalanceRequest) (*v1.GetBalanceResponse, error) {
	return s.handler.GetBalance(ctx, req)
}
