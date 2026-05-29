package conf

// ClientConf holds the gRPC addresses of upstream services this module calls.
type ClientConf struct {
	CustomerAddr string `json:"customer_addr"`
	LedgerAddr   string `json:"ledger_addr"`
}
