package handler

import (
	"context"
	"testing"

	"customer/internal/data"
	"customer/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
)

func TestUpdateKYCStatus_PendingToVerified(t *testing.T) {
	current := sampleIndividualRecord() // KycStatus = PENDING
	updated := sampleIndividualRecord()
	updated.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_VERIFIED)

	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) {
			return current, nil
		},
		updateKYCStatusFn: func(_ context.Context, _ int64, _ int16) (*data.CustomerRecord, error) {
			return updated, nil
		},
	}
	resp, err := newHandler(repo).UpdateKYCStatus(context.Background(), &v1.UpdateKYCStatusRequest{
		Id:        1,
		NewStatus: commonv1.KYCStatus_KYC_STATUS_VERIFIED,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Customer.KycStatus != commonv1.KYCStatus_KYC_STATUS_VERIFIED {
		t.Errorf("kyc_status: got %s", resp.Customer.KycStatus)
	}
}

func TestUpdateKYCStatus_VerifiedToActive(t *testing.T) {
	current := sampleIndividualRecord()
	current.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_VERIFIED)
	updated := sampleIndividualRecord()
	updated.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_ACTIVE)

	repo := &mockRepo{
		getByIDFn:         func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
		updateKYCStatusFn: func(_ context.Context, _ int64, _ int16) (*data.CustomerRecord, error) { return updated, nil },
	}
	resp, err := newHandler(repo).UpdateKYCStatus(context.Background(), &v1.UpdateKYCStatusRequest{
		Id:        1,
		NewStatus: commonv1.KYCStatus_KYC_STATUS_ACTIVE,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Customer.KycStatus != commonv1.KYCStatus_KYC_STATUS_ACTIVE {
		t.Errorf("expected ACTIVE, got %s", resp.Customer.KycStatus)
	}
}

func TestUpdateKYCStatus_InvalidTransition_PendingToActive(t *testing.T) {
	current := sampleIndividualRecord() // PENDING
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
	}
	_, err := newHandler(repo).UpdateKYCStatus(context.Background(), &v1.UpdateKYCStatusRequest{
		Id:        1,
		NewStatus: commonv1.KYCStatus_KYC_STATUS_ACTIVE, // PENDING→ACTIVE not allowed
	})
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestUpdateKYCStatus_InvalidTransition_ClosedToActive(t *testing.T) {
	current := sampleIndividualRecord()
	current.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_CLOSED)
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
	}
	_, err := newHandler(repo).UpdateKYCStatus(context.Background(), &v1.UpdateKYCStatusRequest{
		Id:        1,
		NewStatus: commonv1.KYCStatus_KYC_STATUS_ACTIVE,
	})
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %s", grpcCode(err))
	}
}

func TestUpdateKYCStatus_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ int64) (*data.CustomerRecord, error) {
			return nil, data.ErrNotFound
		},
	}
	_, err := newHandler(repo).UpdateKYCStatus(context.Background(), &v1.UpdateKYCStatusRequest{
		Id:        999,
		NewStatus: commonv1.KYCStatus_KYC_STATUS_VERIFIED,
	})
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %s", grpcCode(err))
	}
}

func TestUpdateKYCStatus_ActiveToSuspended(t *testing.T) {
	current := sampleIndividualRecord()
	current.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_ACTIVE)
	updated := sampleIndividualRecord()
	updated.Customer.KycStatus = int16(commonv1.KYCStatus_KYC_STATUS_SUSPENDED)

	repo := &mockRepo{
		getByIDFn:         func(_ context.Context, _ int64) (*data.CustomerRecord, error) { return current, nil },
		updateKYCStatusFn: func(_ context.Context, _ int64, _ int16) (*data.CustomerRecord, error) { return updated, nil },
	}
	resp, err := newHandler(repo).UpdateKYCStatus(context.Background(), &v1.UpdateKYCStatusRequest{
		Id:        1,
		NewStatus: commonv1.KYCStatus_KYC_STATUS_SUSPENDED,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = model.Customer{} // ensure model import is used
	if resp.Customer.KycStatus != commonv1.KYCStatus_KYC_STATUS_SUSPENDED {
		t.Errorf("expected SUSPENDED, got %s", resp.Customer.KycStatus)
	}
}
