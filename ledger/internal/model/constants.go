package model

// TxStatus values for ledger_transactions.status
const (
	TxStatusPosted   int16 = 1
	TxStatusRejected int16 = 2
)

// EntryDirection values for ledger_entries.direction
const (
	DirectionDebit  int16 = 1
	DirectionCredit int16 = 2
)
