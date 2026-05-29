package handler

import (
	"context"
	"errors"
	"testing"

	"customer/internal/data"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
)

func TestListCustomers_Success(t *testing.T) {
	records := []*data.CustomerRecord{sampleIndividualRecord(), sampleBusinessRecord()}
	repo := &mockRepo{
		listFn: func(_ context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
			if p.Limit != defaultPageSize {
				t.Errorf("expected default page size %d, got %d", defaultPageSize, p.Limit)
			}
			return records, int64(len(records)), nil
		},
	}
	resp, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Customers) != 2 {
		t.Errorf("expected 2 customers, got %d", len(resp.Customers))
	}
	if resp.Page.TotalSize != 2 {
		t.Errorf("expected total_size=2, got %d", resp.Page.TotalSize)
	}
	if resp.Page.NextPageToken != "" {
		t.Errorf("expected no next page token, got %q", resp.Page.NextPageToken)
	}
}

func TestListCustomers_WithStatusFilter(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
			if p.StatusFilter != int16(commonv1.KYCStatus_KYC_STATUS_PENDING) {
				t.Errorf("expected status filter PENDING, got %d", p.StatusFilter)
			}
			return []*data.CustomerRecord{sampleIndividualRecord()}, 1, nil
		},
	}
	resp, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{
		StatusFilter: commonv1.KYCStatus_KYC_STATUS_PENDING,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Customers) != 1 {
		t.Errorf("expected 1 customer, got %d", len(resp.Customers))
	}
}

func TestListCustomers_WithTypeFilter(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
			if p.TypeFilter != int16(commonv1.CustomerType_CUSTOMER_TYPE_BUSINESS) {
				t.Errorf("expected type filter BUSINESS, got %d", p.TypeFilter)
			}
			return []*data.CustomerRecord{sampleBusinessRecord()}, 1, nil
		},
	}
	_, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{
		TypeFilter: commonv1.CustomerType_CUSTOMER_TYPE_BUSINESS,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCustomers_Pagination_NextTokenSet(t *testing.T) {
	// 25 total records, page_size=10, offset=0 → next token should be "10"
	records := make([]*data.CustomerRecord, 10)
	for i := range records {
		records[i] = sampleIndividualRecord()
	}
	repo := &mockRepo{
		listFn: func(_ context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
			return records, 25, nil
		},
	}
	resp, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{
		Page: &commonv1.PageRequest{PageSize: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Page.NextPageToken != "10" {
		t.Errorf("expected next_page_token=10, got %q", resp.Page.NextPageToken)
	}
}

func TestListCustomers_Pagination_SecondPage(t *testing.T) {
	records := make([]*data.CustomerRecord, 5)
	for i := range records {
		records[i] = sampleIndividualRecord()
	}
	repo := &mockRepo{
		listFn: func(_ context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
			if p.Offset != 10 {
				t.Errorf("expected offset=10, got %d", p.Offset)
			}
			return records, 15, nil // total=15, offset=10, got 5 → no next page
		},
	}
	resp, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{
		Page: &commonv1.PageRequest{PageSize: 10, PageToken: "10"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Page.NextPageToken != "" {
		t.Errorf("expected no next page token, got %q", resp.Page.NextPageToken)
	}
}

func TestListCustomers_Empty(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _ data.ListParams) ([]*data.CustomerRecord, int64, error) {
			return nil, 0, nil
		},
	}
	resp, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Customers) != 0 {
		t.Errorf("expected 0 customers, got %d", len(resp.Customers))
	}
	if resp.Page.TotalSize != 0 {
		t.Errorf("expected total_size=0, got %d", resp.Page.TotalSize)
	}
}

func TestListCustomers_PageSizeCappedAt100(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
			if p.Limit != 100 {
				t.Errorf("expected limit capped at 100, got %d", p.Limit)
			}
			return nil, 0, nil
		},
	}
	newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{ //nolint:errcheck
		Page: &commonv1.PageRequest{PageSize: 9999},
	})
}

func TestListCustomers_InternalError(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _ data.ListParams) ([]*data.CustomerRecord, int64, error) {
			return nil, 0, errors.New("db timeout")
		},
	}
	_, err := newHandler(repo).ListCustomers(context.Background(), &v1.ListCustomersRequest{})
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
