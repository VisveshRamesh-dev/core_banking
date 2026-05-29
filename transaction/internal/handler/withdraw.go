package handler

import (
	"context"
	"errors"

	"transaction/internal/data"
	"transaction/internal/mapper"
	"transaction/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *TransactionHandler) Withdraw(ctx context.Context, req *transactionv1.WithdrawRequest) (*transactionv1.WithdrawResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key is required")
	}
	if req.FromAccountId == 0 {
		return nil, status.Error(codes.InvalidArgument, "from_account_id is required")
	}
	if req.AmountMinor <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount_minor must be positive")
	}
	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	existing, err := h.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		return &transactionv1.WithdrawResponse{Transaction: mapper.TransactionToProto(existing)}, nil
	}

	fromID := req.FromAccountId
	var dstRef *string
	if req.DestinationReference != "" {
		dstRef = &req.DestinationReference
	}
	tx := &model.Transaction{
		Kind:                 model.TxKindWithdrawal,
		State:                model.TxStatePending,
		FromAccountID:        &fromID,
		AmountMinor:          req.AmountMinor,
		Currency:             req.Currency,
		IdempotencyKey:       req.IdempotencyKey,
		DestinationReference: dstRef,
	}
	if err := h.repo.Create(ctx, tx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := h.validateAccountActive(ctx, req.FromAccountId); err != nil {
		_ = h.repo.MarkFailed(ctx, tx.ID, err.Error())
		return nil, err
	}

	// source account DEBIT, settlement account CREDIT — transaction stays balanced.
	ledgerResp, err := h.ledger.PostTransaction(ctx, &ledgerv1.PostTransactionRequest{
		IdempotencyKey: req.IdempotencyKey + ":ledger",
		Description:    "Withdrawal: " + req.DestinationReference,
		Entries: []*ledgerv1.Entry{
			{AccountId: req.FromAccountId, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT, AmountMinor: req.AmountMinor, Currency: req.Currency},
			{AccountId: h.settlementAcctID, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT, AmountMinor: req.AmountMinor, Currency: req.Currency},
		},
	})
	if err != nil {
		reason := err.Error()
		_ = h.repo.MarkFailed(ctx, tx.ID, reason)
		return nil, status.Error(codes.Internal, reason)
	}

	ledgerTxID := ledgerResp.Transaction.Id
	if err := h.repo.MarkCompleted(ctx, tx.ID, ledgerTxID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	tx.State = model.TxStateCompleted
	tx.LedgerTransactionID = &ledgerTxID

	return &transactionv1.WithdrawResponse{Transaction: mapper.TransactionToProto(tx)}, nil
}
