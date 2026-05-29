package mapper

import (
	"transaction/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TransactionToProto(t *model.Transaction) *transactionv1.Transaction {
	tx := &transactionv1.Transaction{
		Id:             t.ID,
		Kind:           commonv1.TransactionKind(t.Kind),
		State:          commonv1.TransactionState(t.State),
		Amount:         &commonv1.Money{AmountMinor: t.AmountMinor, Currency: t.Currency},
		IdempotencyKey: t.IdempotencyKey,
		CreatedAt:      timestamppb.New(t.CreatedAt),
	}
	if t.FromAccountID != nil {
		tx.FromAccountId = t.FromAccountID
	}
	if t.ToAccountID != nil {
		tx.ToAccountId = t.ToAccountID
	}
	if t.LedgerTransactionID != nil {
		tx.LedgerTransactionId = t.LedgerTransactionID
	}
	if t.FailureReason != nil {
		tx.FailureReason = *t.FailureReason
	}
	if t.CompletedAt != nil {
		tx.CompletedAt = timestamppb.New(*t.CompletedAt)
	}
	return tx
}
