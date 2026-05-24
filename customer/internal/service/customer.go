package service

import (
	"context"

	v1 "github.com/visvesh-ramesh/corebank/v1/customer"

	"customer/internal/handler"
)

// CustomerService implements the CustomerServiceServer gRPC/HTTP interface.
// It holds a *data.Data for direct database access — no intermediate biz layer.
type CustomerService struct {
	v1.UnimplementedCustomerServiceServer
	handler *handler.CustomerHandler
}

func NewCustomerService(handler *handler.CustomerHandler) *CustomerService {
	return &CustomerService{handler: handler}
}

// Onboard creates a new customer record in PENDING/KYC_PENDING status.
// POST /v1/customers
func (s *CustomerService) Onboard(ctx context.Context, req *v1.OnboardRequest) (*v1.OnboardResponse, error) {
	// TODO: validate req, persist customer, return created record
	return s.handler.Onboard(ctx, req)
}

// GetCustomer retrieves a customer by ID.
// GET /v1/customers/{id}
func (s *CustomerService) GetCustomer(ctx context.Context, req *v1.GetCustomerRequest) (*v1.GetCustomerResponse, error) {
	// TODO: query customer by req.Id, return record
	return &v1.GetCustomerResponse{}, nil
}

// UpdateKYCStatus transitions the customer through the KYC state machine.
// PATCH /v1/customers/{id}/kyc-status
func (s *CustomerService) UpdateKYCStatus(ctx context.Context, req *v1.UpdateKYCStatusRequest) (*v1.UpdateKYCStatusResponse, error) {
	// TODO: validate transition, update kyc_status, persist audit note
	return &v1.UpdateKYCStatusResponse{}, nil
}

// CloseCustomer moves the customer to the terminal CLOSED status.
// POST /v1/customers/{id}:close
func (s *CustomerService) CloseCustomer(ctx context.Context, req *v1.CloseCustomerRequest) (*v1.CloseCustomerResponse, error) {
	// TODO: validate not already closed, set status CLOSED
	return &v1.CloseCustomerResponse{}, nil
}

// ListCustomers returns a paginated list filtered by KYC status and/or customer type.
// GET /v1/customers
func (s *CustomerService) ListCustomers(ctx context.Context, req *v1.ListCustomersRequest) (*v1.ListCustomersResponse, error) {
	// TODO: apply status_filter, type_filter, page params, return results
	return &v1.ListCustomersResponse{}, nil
}
