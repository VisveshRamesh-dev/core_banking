package handler

import (
	"context"
	"errors"
	"testing"

	"customer/internal/data"

	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
)

func TestGetCustomer_Success(t *testing.T) {
	want := sampleIndividualRecord()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, id int64) (*data.CustomerRecord, error) {
			if id != 1 {
				t.Errorf("expected id=1, got %d", id)
			}
			return want, nil
		},
	}
	resp, err := newHandler(repo).GetCustomer(context.Background(), &v1.GetCustomerRequest{Id: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Customer.Id != 1 {
		t.Errorf("customer id: got %d want 1", resp.Customer.Id)
	}
	if resp.Customer.Email != "arjun@example.com" {
		t.Errorf("email: got %q", resp.Customer.Email)
	}
}

func TestGetCustomer_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) {
			return nil, data.ErrNotFound
		},
	}
	_, err := newHandler(repo).GetCustomer(context.Background(), &v1.GetCustomerRequest{Id: 999})
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestGetCustomer_InternalError(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) {
			return nil, errors.New("connection refused")
		},
	}
	_, err := newHandler(repo).GetCustomer(context.Background(), &v1.GetCustomerRequest{Id: 1})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
