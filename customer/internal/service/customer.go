package service

import (
	"context"

	v1 "github.com/visvesh-ramesh/corebank/v1/customer"

	"customer/internal/handler"
)

// CustomerService is the thin gRPC/HTTP adapter. All logic lives in handler.
type CustomerService struct {
	v1.UnimplementedCustomerServiceServer
	handler *handler.CustomerHandler
}

func NewCustomerService(handler *handler.CustomerHandler) *CustomerService {
	return &CustomerService{handler: handler}
}

func (s *CustomerService) Onboard(ctx context.Context, req *v1.OnboardRequest) (*v1.OnboardResponse, error) {
	return s.handler.Onboard(ctx, req)
}

func (s *CustomerService) GetCustomer(ctx context.Context, req *v1.GetCustomerRequest) (*v1.GetCustomerResponse, error) {
	return s.handler.GetCustomer(ctx, req)
}

func (s *CustomerService) UpdateKYCStatus(ctx context.Context, req *v1.UpdateKYCStatusRequest) (*v1.UpdateKYCStatusResponse, error) {
	return s.handler.UpdateKYCStatus(ctx, req)
}

func (s *CustomerService) CloseCustomer(ctx context.Context, req *v1.CloseCustomerRequest) (*v1.CloseCustomerResponse, error) {
	return s.handler.CloseCustomer(ctx, req)
}

func (s *CustomerService) ListCustomers(ctx context.Context, req *v1.ListCustomersRequest) (*v1.ListCustomersResponse, error) {
	return s.handler.ListCustomers(ctx, req)
}
