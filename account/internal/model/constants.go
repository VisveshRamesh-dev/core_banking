package model

const (
	AccountStatusPending int16 = 1
	AccountStatusActive  int16 = 2
	AccountStatusFrozen  int16 = 3
	AccountStatusClosed  int16 = 4

	AccountTypeSavings int16 = 1
	AccountTypeCurrent int16 = 2
	AccountTypeWallet  int16 = 3
	AccountTypeLoan    int16 = 4
)
