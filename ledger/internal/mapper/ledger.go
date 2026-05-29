package mapper

import (
	"ledger/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TxRecordToProto converts a TxRecord into a LedgerTransaction proto.
func TxRecordToProto(tx model.LedgerTransaction, entries []model.LedgerEntry) *v1.LedgerTransaction {
	protoEntries := make([]*v1.Entry, len(entries))
	for i, e := range entries {
		protoEntries[i] = entryToProto(e)
	}

	t := &v1.LedgerTransaction{
		Id:             tx.ID,
		Entries:        protoEntries,
		Status:         commonv1.LedgerTransactionStatus(tx.Status),
		Description:    derefString(tx.Description),
		IdempotencyKey: tx.IdempotencyKey,
		PostedAt:       timestamppb.New(tx.PostedAt),
	}
	if tx.ReversesTransactionID != nil {
		t.ReversesTransactionId = tx.ReversesTransactionID
	}
	return t
}

// PostRequestToModels converts a PostTransactionRequest into DB rows.
// Entries have TransactionID=0; the repo fills it after insert.
func PostRequestToModels(req *v1.PostTransactionRequest) (model.LedgerTransaction, []model.LedgerEntry) {
	desc := req.Description
	tx := model.LedgerTransaction{
		Status:         model.TxStatusPosted,
		IdempotencyKey: req.IdempotencyKey,
		Description:    nilIfEmpty(desc),
	}
	entries := make([]model.LedgerEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = model.LedgerEntry{
			AccountID:   e.AccountId,
			Direction:   int16(e.Direction),
			AmountMinor: e.AmountMinor,
			Currency:    e.Currency,
		}
	}
	return tx, entries
}

// ReversalEntries builds the mirror entries for a reversal transaction.
func ReversalEntries(original []model.LedgerEntry) []model.LedgerEntry {
	out := make([]model.LedgerEntry, len(original))
	for i, e := range original {
		flipped := model.DirectionCredit
		if e.Direction == model.DirectionCredit {
			flipped = model.DirectionDebit
		}
		out[i] = model.LedgerEntry{
			AccountID:   e.AccountID,
			Direction:   flipped,
			AmountMinor: e.AmountMinor,
			Currency:    e.Currency,
		}
	}
	return out
}

// ── helpers ──────────────────────────────────────────────────────────────────

func entryToProto(e model.LedgerEntry) *v1.Entry {
	return &v1.Entry{
		Id:          e.ID,
		AccountId:   e.AccountID,
		Direction:   commonv1.EntryDirection(e.Direction),
		AmountMinor: e.AmountMinor,
		Currency:    e.Currency,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
