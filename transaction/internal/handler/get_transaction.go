package handler

import (
	"context"
	"errors"

	"transaction/internal/data"
	"transaction/internal/mapper"

	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *TransactionHandler) GetTransaction(ctx context.Context, req *transactionv1.GetTransactionRequest) (*transactionv1.GetTransactionResponse, error) {
	tx, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &transactionv1.GetTransactionResponse{Transaction: mapper.TransactionToProto(tx)}, nil
}
