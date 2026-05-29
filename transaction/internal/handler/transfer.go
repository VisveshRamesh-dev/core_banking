package handler

import (
	"context"
	"errors"

	"transaction/internal/data"
	"transaction/internal/mapper"
	"transaction/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *TransactionHandler) Transfer(ctx context.Context, req *transactionv1.TransferRequest) (*transactionv1.TransferResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key is required")
	}
	if req.FromAccountId == 0 {
		return nil, status.Error(codes.InvalidArgument, "from_account_id is required")
	}
	if req.ToAccountId == 0 {
		return nil, status.Error(codes.InvalidArgument, "to_account_id is required")
	}
	if req.FromAccountId == req.ToAccountId {
		return nil, status.Error(codes.InvalidArgument, "from_account_id and to_account_id must differ")
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
		return &transactionv1.TransferResponse{Transaction: mapper.TransactionToProto(existing)}, nil
	}

	fromID := req.FromAccountId
	toID := req.ToAccountId
	tx := &model.Transaction{
		Kind:          model.TxKindTransfer,
		State:         model.TxStatePending,
		FromAccountID: &fromID,
		ToAccountID:   &toID,
		AmountMinor:   req.AmountMinor,
		Currency:      req.Currency,
		IdempotencyKey: req.IdempotencyKey,
	}
	if err := h.repo.Create(ctx, tx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Validate both accounts are ACTIVE.
	if err := h.validateAccountActive(ctx, req.FromAccountId); err != nil {
		_ = h.repo.MarkFailed(ctx, tx.ID, err.Error())
		return nil, err
	}
	if err := h.validateAccountActive(ctx, req.ToAccountId); err != nil {
		_ = h.repo.MarkFailed(ctx, tx.ID, err.Error())
		return nil, err
	}

	// Post balanced debit/credit to ledger.
	ledgerResp, err := h.ledger.PostTransaction(ctx, &ledgerv1.PostTransactionRequest{
		IdempotencyKey: req.IdempotencyKey + ":ledger",
		Description:    req.Description,
		Entries: []*ledgerv1.Entry{
			{AccountId: req.FromAccountId, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_DEBIT, AmountMinor: req.AmountMinor, Currency: req.Currency},
			{AccountId: req.ToAccountId, Direction: commonv1.EntryDirection_ENTRY_DIRECTION_CREDIT, AmountMinor: req.AmountMinor, Currency: req.Currency},
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

	return &transactionv1.TransferResponse{Transaction: mapper.TransactionToProto(tx)}, nil
}

// validateAccountActive fetches an account and returns an error if it is not ACTIVE.
func (h *TransactionHandler) validateAccountActive(ctx context.Context, accountID int64) error {
	resp, err := h.accounts.GetAccount(ctx, &accountv1.GetAccountRequest{Id: accountID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return status.Errorf(codes.FailedPrecondition, "account %d not found", accountID)
		}
		return status.Error(codes.Internal, err.Error())
	}
	if resp.Account.Status != commonv1.AccountStatus_ACCOUNT_STATUS_ACTIVE {
		return status.Errorf(codes.FailedPrecondition,
			"account %d is not ACTIVE (status: %s)", accountID, resp.Account.Status)
	}
	return nil
}
