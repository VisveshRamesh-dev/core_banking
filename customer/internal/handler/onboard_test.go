package handler

import (
	"context"
	"errors"
	"testing"

	"customer/internal/data"

	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
)

func TestOnboard_IndividualSuccess(t *testing.T) {
	want := sampleIndividualRecord()
	repo := &mockRepo{
		createFn: func(_ context.Context, rec *data.CustomerRecord) error {
			rec.Customer.ID = 1
			rec.Customer.IndividualID = ptr(int64(1))
			if rec.Individual != nil {
				rec.Individual.ID = 1
			}
			return nil
		},
	}
	h := newHandler(repo)
	resp, err := h.Onboard(context.Background(), sampleOnboardIndividualReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Customer == nil {
		t.Fatal("expected customer in response")
	}
	if resp.Customer.Email != want.Customer.Email {
		t.Errorf("email: got %q want %q", resp.Customer.Email, want.Customer.Email)
	}
	if resp.Customer.GetIndividual() == nil {
		t.Error("expected individual details in response")
	}
	if len(resp.Customer.Phones) == 0 {
		t.Error("expected at least one phone")
	}
}

func TestOnboard_BusinessSuccess(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, rec *data.CustomerRecord) error {
			rec.Customer.ID = 2
			rec.Customer.BusinessID = ptr(int64(1))
			if rec.Business != nil {
				rec.Business.ID = 1
			}
			return nil
		},
	}
	h := newHandler(repo)
	resp, err := h.Onboard(context.Background(), sampleOnboardBusinessReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Customer.GetBusiness() == nil {
		t.Error("expected business details in response")
	}
	biz := resp.Customer.GetBusiness()
	if biz.CompanyName != "Acme Corp Pvt Ltd" {
		t.Errorf("company_name: got %q", biz.CompanyName)
	}
	if biz.Proprietor == nil {
		t.Error("expected proprietor in business details")
	}
}

func TestOnboard_MissingFirstName(t *testing.T) {
	req := sampleOnboardIndividualReq()
	req.FirstName = ""
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_MissingEmail(t *testing.T) {
	req := sampleOnboardIndividualReq()
	req.Email = ""
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_NoDetails(t *testing.T) {
	req := &v1.OnboardRequest{FirstName: "A", LastName: "B", Email: "a@b.com"}
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_IndividualMissingPhone(t *testing.T) {
	req := sampleOnboardIndividualReq()
	req.Phones = nil
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_IndividualMissingAddress(t *testing.T) {
	req := sampleOnboardIndividualReq()
	req.Addresses = nil
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_BusinessMissingCompanyPhone(t *testing.T) {
	req := sampleOnboardBusinessReq()
	req.GetBusiness().CompanyPhones = nil
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_BusinessMissingProprietor(t *testing.T) {
	req := sampleOnboardBusinessReq()
	req.GetBusiness().Proprietor = nil
	_, err := newHandler(&mockRepo{}).Onboard(context.Background(), req)
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", grpcCode(err))
	}
}

func TestOnboard_RepoError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *data.CustomerRecord) error {
			return errors.New("db down")
		},
	}
	_, err := newHandler(repo).Onboard(context.Background(), sampleOnboardIndividualReq())
	if grpcCode(err) != codes.Internal {
		t.Errorf("expected Internal, got %s", grpcCode(err))
	}
}
