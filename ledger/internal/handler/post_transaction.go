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

func (h *LedgerHandler) PostTransaction(ctx context.Context, req *v1.PostTransactionRequest) (*v1.PostTransactionResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key is required")
	}
	if len(req.Entries) < 2 {
		return nil, status.Error(codes.InvalidArgument, "at least two entries are required")
	}
	if err := validateEntries(req.Entries); err != nil {
		return nil, err
	}

	// Idempotency: return existing transaction if key already used.
	existing, err := h.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		proto := mapper.TxRecordToProto(existing.Tx, existing.Entries)
		return &v1.PostTransactionResponse{Transaction: proto}, nil
	}

	tx, entries := mapper.PostRequestToModels(req)
	rec := &data.TxRecord{Tx: tx, Entries: entries}

	if err := h.repo.Post(ctx, rec); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.TxRecordToProto(rec.Tx, rec.Entries)
	return &v1.PostTransactionResponse{Transaction: proto}, nil
}

// validateEntries enforces the double-entry invariant:
//   - every entry amount must be positive
//   - all entries must share one currency
//   - signed sum (credit - debit) must equal zero
func validateEntries(entries []*v1.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	currency := entries[0].Currency
	var sum int64
	for _, e := range entries {
		if e.AmountMinor <= 0 {
			return status.Error(codes.InvalidArgument, "entry amount_minor must be positive")
		}
		if e.AccountId == 0 {
			return status.Error(codes.InvalidArgument, "entry account_id is required")
		}
		if e.Currency != currency {
			return status.Error(codes.InvalidArgument, "all entries must share the same currency")
		}
		if e.Direction == commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT {
			sum += e.AmountMinor
		} else {
			sum -= e.AmountMinor
		}
	}
	if sum != 0 {
		return status.Errorf(codes.InvalidArgument,
			"entries do not balance: net signed sum is %d (must be zero)", sum)
	}
	return nil
}

// directionFromModel is used in reversal — unexported helper shared across files.
func flipDirection(d int16) int16 {
	if d == model.DirectionCredit {
		return model.DirectionDebit
	}
	return model.DirectionCredit
}
