package handler

import (
	"context"
	"testing"

	"customer/internal/data"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
)

func TestCloseCustomer_ActiveToClosed(t *testing.T) {
	current := sampleIndividualRecord()
	current.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_ACTIVE)
	closed := sampleIndividualRecord()
	closed.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_CLOSED)

	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
		updateKYCStatusFn: func(_ context.Context, _ int64, s int16) (*data.CustomerRecord, error) {
			if s != int16(commonv1.KYCStatus_KYC_STATUS_CLOSED) {
				t.Errorf("expected CLOSED status, got %d", s)
			}
			return closed, nil
		},
	}
	resp, err := newHandler(repo).CloseCustomer(context.Background(), &v1.CloseCustomerRequest{
		Id:     1,
		Reason: "customer requested closure",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Customer.KycStatus != commonv1.KYCStatus_KYC_STATUS_CLOSED {
		t.Errorf("expected CLOSED, got %s", resp.Customer.KycStatus)
	}
}

func TestCloseCustomer_AlreadyClosed(t *testing.T) {
	current := sampleIndividualRecord()
	current.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_CLOSED)
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
	}
	_, err := newHandler(repo).CloseCustomer(context.Background(), &v1.CloseCustomerRequest{Id: 1})
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestCloseCustomer_PendingCannotClose(t *testing.T) {
	current := sampleIndividualRecord() // KycStatus = PENDING
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
	}
	_, err := newHandler(repo).CloseCustomer(context.Background(), &v1.CloseCustomerRequest{Id: 1})
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition for PENDING→CLOSED, got %s", grpcCode(err))
	}
}

func TestCloseCustomer_SuspendedCannotClose(t *testing.T) {
	current := sampleIndividualRecord()
	current.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_SUSPENDED)
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
	}
	_, err := newHandler(repo).CloseCustomer(context.Background(), &v1.CloseCustomerRequest{Id: 1})
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition for SUSPENDED→CLOSED, got %s", grpcCode(err))
	}
}

func TestCloseCustomer_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) {
			return nil, data.ErrNotFound
		},
	}
	_, err := newHandler(repo).CloseCustomer(context.Background(), &v1.CloseCustomerRequest{Id: 999})
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}
