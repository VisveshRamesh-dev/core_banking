package model

const (
	TxKindTransfer   int16 = 1
	TxKindDeposit    int16 = 2
	TxKindWithdrawal int16 = 3
	TxKindReversal   int16 = 4

	TxStatePending   int16 = 1
	TxStateCompleted int16 = 2
	TxStateFailed    int16 = 3
)
