package handler

import (
	"context"
	"time"

	"customer/internal/data"
	"customer/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ── mock repo ────────────────────────────────────────────────────────────────

type mockRepo struct {
	createFn          func(ctx context.Context, rec *data.CustomerRecord) error
	getByIDFn         func(ctx context.Context, id int64) (*data.CustomerRecord, error)
	updateKYCStatusFn func(ctx context.Context, id int64, newStatus int16) (*data.CustomerRecord, error)
	listFn            func(ctx context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error)
}

func (m *mockRepo) Create(ctx context.Context, rec *data.CustomerRecord) error {
	return m.createFn(ctx, rec)
}
func (m *mockRepo) GetByID(ctx context.Context, id int64) (*data.CustomerRecord, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) UpdateKYCStatus(ctx context.Context, id int64, newStatus int16) (*data.CustomerRecord, error) {
	return m.updateKYCStatusFn(ctx, id, newStatus)
}
func (m *mockRepo) List(ctx context.Context, p data.ListParams) ([]*data.CustomerRecord, int64, error) {
	return m.listFn(ctx, p)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func newHandler(repo customerRepo) *CustomerHandler {
	return &CustomerHandler{customerRepo: repo}
}

func ptr[T any](v T) *T { return &v }

func sampleIndividualRecord() *data.CustomerRecord {
	return &data.CustomerRecord{
		Customer: model.Customer{
			ID:           1,
			CustomerType: int16(commonv1.CustomerType_CUSTOMER_TYPE_INDIVIDUAL),
			FirstName:    "Arjun",
			LastName:     "Sharma",
			FullName:     ptr("Arjun Sharma"),
			Email:        "arjun@example.com",
			KycStatus:    int16(commonv1.KYCStatus_KYC_STATUS_PENDING),
			IndividualID: ptr(int64(1)),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		Individual: &model.IndividualCustomer{
			ID:         1,
			NationalID: ptr("AAAA1234B"),
			Nationality: ptr("IN"),
		},
		Phones: []model.Phone{
			{ID: 1, PhoneType: int16(commonv1.PhoneType_PHONE_TYPE_MOBILE), CountryCode: "+91", Number: "9876543210", IsPrimary: true},
		},
		Addresses: []model.Address{
			{ID: 1, AddressType: int16(commonv1.AddressType_ADDRESS_TYPE_HOME), Line1: "42 MG Road", City: "Bengaluru", State: "Karnataka", PostalCode: "560001", Country: "IN", IsPrimary: true},
		},
	}
}

func sampleBusinessRecord() *data.CustomerRecord {
	return &data.CustomerRecord{
		Customer: model.Customer{
			ID:           2,
			CustomerType: int16(commonv1.CustomerType_CUSTOMER_TYPE_BUSINESS),
			FirstName:    "Priya",
			LastName:     "Nair",
			FullName:     ptr("Acme Corp Pvt Ltd"),
			Email:        "contact@acme.in",
			KycStatus:    int16(commonv1.KYCStatus_KYC_STATUS_PENDING),
			BusinessID:   ptr(int64(1)),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		Business: &model.BusinessCustomer{
			ID:            1,
			CompanyName:   "Acme Corp Pvt Ltd",
			PropFirstName: "Priya",
			PropLastName:  "Nair",
			PropEmail:     "priya@acme.in",
		},
		BizPhones: []model.Phone{
			{ID: 2, PhoneType: int16(commonv1.PhoneType_PHONE_TYPE_WORK), CountryCode: "+91", Number: "2222345678", IsPrimary: true},
		},
		BizAddrs: []model.Address{
			{ID: 2, AddressType: int16(commonv1.AddressType_ADDRESS_TYPE_REGISTERED), Line1: "501 Business Park", City: "Mumbai", State: "Maharashtra", PostalCode: "400001", Country: "IN", IsPrimary: true},
		},
		PropPhones: []model.Phone{
			{ID: 3, PhoneType: int16(commonv1.PhoneType_PHONE_TYPE_MOBILE), CountryCode: "+91", Number: "9988776655", IsPrimary: true},
		},
	}
}

func sampleOnboardIndividualReq() *v1.OnboardRequest {
	return &v1.OnboardRequest{
		FirstName: "Arjun",
		LastName:  "Sharma",
		Email:     "arjun@example.com",
		Phones: []*commonv1.Phone{
			{Type: commonv1.PhoneType_PHONE_TYPE_MOBILE, CountryCode: "+91", Number: "9876543210", IsPrimary: true},
		},
		Addresses: []*commonv1.Address{
			{Type: commonv1.AddressType_ADDRESS_TYPE_HOME, Line1: "42 MG Road", City: "Bengaluru", State: "Karnataka", PostalCode: "560001", Country: "IN", IsPrimary: true},
		},
		Details: &v1.OnboardRequest_Individual{
			Individual: &v1.IndividualDetails{
				DateOfBirth: "1990-06-15",
				Nationality: "IN",
				NationalId:  "AAAA1234B",
			},
		},
	}
}

func sampleOnboardBusinessReq() *v1.OnboardRequest {
	return &v1.OnboardRequest{
		FirstName: "Priya",
		LastName:  "Nair",
		Email:     "contact@acme.in",
		Details: &v1.OnboardRequest_Business{
			Business: &v1.BusinessDetails{
				CompanyName:        "Acme Corp Pvt Ltd",
				RegistrationNumber: "U12345MH2020PTC123456",
				TaxId:              "27AAACA1234A1Z5",
				CompanyPhones: []*commonv1.Phone{
					{Type: commonv1.PhoneType_PHONE_TYPE_WORK, CountryCode: "+91", Number: "2222345678", IsPrimary: true},
				},
				RegisteredAddresses: []*commonv1.Address{
					{Type: commonv1.AddressType_ADDRESS_TYPE_REGISTERED, Line1: "501 Business Park", City: "Mumbai", State: "Maharashtra", PostalCode: "400001", Country: "IN", IsPrimary: true},
				},
				Proprietor: &v1.ProprietorInfo{
					FirstName:  "Priya",
					LastName:   "Nair",
					Email:      "priya@acme.in",
					NationalId: "BBBB5678C",
					Phones: []*commonv1.Phone{
						{Type: commonv1.PhoneType_PHONE_TYPE_MOBILE, CountryCode: "+91", Number: "9988776655", IsPrimary: true},
					},
				},
			},
		},
	}
}

// grpcCode extracts the gRPC status code from an error.
func grpcCode(err error) codes.Code {
	if s, ok := status.FromError(err); ok {
		return s.Code()
	}
	return codes.Unknown
}
