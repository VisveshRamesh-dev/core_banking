package data

import (
	"context"

	"transaction/internal/conf"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AccountClient is the subset of AccountServiceClient the handler needs.
type AccountClient interface {
	GetAccount(ctx context.Context, in *accountv1.GetAccountRequest, opts ...grpc.CallOption) (*accountv1.GetAccountResponse, error)
}

// LedgerClient is the subset of LedgerServiceClient the handler needs.
type LedgerClient interface {
	PostTransaction(ctx context.Context, in *ledgerv1.PostTransactionRequest, opts ...grpc.CallOption) (*ledgerv1.PostTransactionResponse, error)
}

// NewAccountGRPCClient dials the account service and returns a client.
func NewAccountGRPCClient(c *conf.ClientConf) (AccountClient, func(), error) {
	conn, err := grpc.NewClient(c.AccountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return accountv1.NewAccountServiceClient(conn), func() { conn.Close() }, nil
}

// NewLedgerGRPCClient dials the ledger service and returns a client.
func NewLedgerGRPCClient(c *conf.ClientConf) (LedgerClient, func(), error) {
	conn, err := grpc.NewClient(c.LedgerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return ledgerv1.NewLedgerServiceClient(conn), func() { conn.Close() }, nil
}
