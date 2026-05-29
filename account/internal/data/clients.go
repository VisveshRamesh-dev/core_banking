package data

import (
	"context"

	"account/internal/conf"

	customerv1 "github.com/visvesh-ramesh/corebank/v1/customer"
	ledgerv1 "github.com/visvesh-ramesh/corebank/v1/ledger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CustomerClient is the subset of CustomerServiceClient the handler needs.
type CustomerClient interface {
	GetCustomer(ctx context.Context, in *customerv1.GetCustomerRequest, opts ...grpc.CallOption) (*customerv1.GetCustomerResponse, error)
}

// LedgerClient is the subset of LedgerServiceClient the handler needs.
type LedgerClient interface {
	GetBalance(ctx context.Context, in *ledgerv1.GetBalanceRequest, opts ...grpc.CallOption) (*ledgerv1.GetBalanceResponse, error)
}

// NewCustomerGRPCClient dials the customer service and returns a client.
func NewCustomerGRPCClient(c *conf.ClientConf) (CustomerClient, func(), error) {
	conn, err := grpc.NewClient(c.CustomerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return customerv1.NewCustomerServiceClient(conn), func() { conn.Close() }, nil
}

// NewLedgerGRPCClient dials the ledger service and returns a client.
func NewLedgerGRPCClient(c *conf.ClientConf) (LedgerClient, func(), error) {
	conn, err := grpc.NewClient(c.LedgerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return ledgerv1.NewLedgerServiceClient(conn), func() { conn.Close() }, nil
}
