package service

import (
	"context"

	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"github.com/google/wire"

	"transaction/internal/handler"
)

var ProviderSet = wire.NewSet(NewTransactionService)

type TransactionService struct {
	transactionv1.UnimplementedTransactionServiceServer
	handler *handler.TransactionHandler
}

func NewTransactionService(h *handler.TransactionHandler) *TransactionService {
	return &TransactionService{handler: h}
}

func (s *TransactionService) Transfer(ctx context.Context, req *transactionv1.TransferRequest) (*transactionv1.TransferResponse, error) {
	return s.handler.Transfer(ctx, req)
}

func (s *TransactionService) Deposit(ctx context.Context, req *transactionv1.DepositRequest) (*transactionv1.DepositResponse, error) {
	return s.handler.Deposit(ctx, req)
}

func (s *TransactionService) Withdraw(ctx context.Context, req *transactionv1.WithdrawRequest) (*transactionv1.WithdrawResponse, error) {
	return s.handler.Withdraw(ctx, req)
}

func (s *TransactionService) GetTransaction(ctx context.Context, req *transactionv1.GetTransactionRequest) (*transactionv1.GetTransactionResponse, error) {
	return s.handler.GetTransaction(ctx, req)
}

func (s *TransactionService) ListAccountTransactions(ctx context.Context, req *transactionv1.ListAccountTransactionsRequest) (*transactionv1.ListAccountTransactionsResponse, error) {
	return s.handler.ListAccountTransactions(ctx, req)
}
