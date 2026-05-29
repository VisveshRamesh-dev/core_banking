package handler

import (
	"context"
	"errors"

	"ledger/internal/data"
	"ledger/internal/mapper"

	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *LedgerHandler) GetTransaction(ctx context.Context, req *v1.GetTransactionRequest) (*v1.GetTransactionResponse, error) {
	rec, err := h.repo.GetByID(ctx, req.TransactionId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction %d not found", req.TransactionId)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.TxRecordToProto(rec.Tx, rec.Entries)
	return &v1.GetTransactionResponse{Transaction: proto}, nil
}
