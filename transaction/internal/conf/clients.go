package conf

// ClientConf holds gRPC addresses of upstream services.
type ClientConf struct {
	AccountAddr string `json:"account_addr"`
	LedgerAddr  string `json:"ledger_addr"`
}

// AppConf holds application-level settings.
type AppConf struct {
	// SettlementAccountID is the ledger account used as counterpart for
	// deposits (external → account) and withdrawals (account → external).
	SettlementAccountID int64 `json:"settlement_account_id"`
}
