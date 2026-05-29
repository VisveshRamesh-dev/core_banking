package handler

import (
	"context"
	"errors"

	"ledger/internal/data"
	"ledger/internal/mapper"
	"ledger/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *LedgerHandler) ReverseTransaction(ctx context.Context, req *v1.ReverseTransactionRequest) (*v1.ReverseTransactionResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key is required")
	}

	// Idempotency: return existing reversal if already done.
	existing, err := h.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		proto := mapper.TxRecordToProto(existing.Tx, existing.Entries)
		return &v1.ReverseTransactionResponse{Transaction: proto}, nil
	}

	// Load the original transaction.
	original, err := h.repo.GetByID(ctx, req.TransactionId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction %d not found", req.TransactionId)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if original.Tx.Status != model.TxStatusPosted {
		return nil, status.Error(codes.FailedPrecondition, "only POSTED transactions can be reversed")
	}

	// Build the reversing transaction.
	desc := "Reversal: " + req.Reason
	origID := original.Tx.ID
	reversalTx := model.LedgerTransaction{
		Status:                model.TxStatusPosted,
		IdempotencyKey:        req.IdempotencyKey,
		Description:           &desc,
		ReversesTransactionID: &origID,
	}
	reversalEntries := mapper.ReversalEntries(original.Entries)
	rec := &data.TxRecord{Tx: reversalTx, Entries: reversalEntries}

	if err := h.repo.Post(ctx, rec); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.TxRecordToProto(rec.Tx, rec.Entries)
	return &v1.ReverseTransactionResponse{Transaction: proto}, nil
}

// Ensure commonv1 import is used (referenced in post_transaction.go via same package)
var _ = commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT
