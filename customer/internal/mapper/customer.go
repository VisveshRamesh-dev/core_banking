package mapper

import (
	"fmt"
	"time"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/protobuf/types/known/timestamppb"

	"customer/internal/model"
)

// CustomerToProto builds a full v1.Customer proto from the database aggregate.
//
// For INDIVIDUAL:  phones/addresses are the individual's personal contacts.
// For BUSINESS:    phones/addresses = company contacts; proprietor contacts
//                  live inside BusinessDetails.Proprietor.Phones.
func CustomerToProto(
	c model.Customer,
	individual *model.IndividualCustomer,
	business *model.BusinessCustomer,
	phones []model.Phone,
	addresses []model.Address,
	bizPhones []model.Phone,
	bizAddrs []model.Address,
	propPhones []model.Phone,
) *v1.Customer {
	cust := &v1.Customer{
		Id:           c.ID,
		CustomerType: commonv1.CustomerType(c.CustomerType),
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		FullName:     derefString(c.FullName),
		Email:        c.Email,
		KycStatus:    commonv1.KYCStatus(c.KycStatus),
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
	}

	switch {
	case individual != nil:
		cust.Phones = phonesToProto(phones)
		cust.Addresses = addressesToProto(addresses)

		ind := &v1.IndividualDetails{
			NationalId:  derefString(individual.NationalID),
			Nationality: derefString(individual.Nationality),
		}
		if individual.DateOfBirth != nil {
			ind.DateOfBirth = individual.DateOfBirth.Format("2006-01-02")
		}
		cust.Details = &v1.Customer_Individual{Individual: ind}

	case business != nil:
		// For a business customer, top-level phones/addresses mirror company contacts.
		cust.Phones = phonesToProto(bizPhones)
		cust.Addresses = addressesToProto(bizAddrs)

		prop := &v1.ProprietorInfo{
			FirstName:  business.PropFirstName,
			LastName:   business.PropLastName,
			Email:      business.PropEmail,
			NationalId: derefString(business.PropNationalID),
			Phones:     phonesToProto(propPhones),
		}
		biz := &v1.BusinessDetails{
			CompanyName:         business.CompanyName,
			RegistrationNumber:  derefString(business.RegistrationNumber),
			TaxId:               derefString(business.TaxID),
			CompanyPhones:       phonesToProto(bizPhones),
			RegisteredAddresses: addressesToProto(bizAddrs),
			Proprietor:          prop,
		}
		cust.Details = &v1.Customer_Business{Business: biz}
	}

	return cust
}

// OnboardRequestToModels converts an OnboardRequest into all DB rows needed
// to persist the customer aggregate. IDs are left zero; the repo fills them.
func OnboardRequestToModels(req *v1.OnboardRequest) (
	cust model.Customer,
	individual *model.IndividualCustomer,
	business *model.BusinessCustomer,
	phones []model.Phone,
	addresses []model.Address,
	bizPhones []model.Phone,
	bizAddrs []model.Address,
	propPhones []model.Phone,
) {
	fullName := req.FullName
	if fullName == "" {
		fullName = fmt.Sprintf("%s %s", req.FirstName, req.LastName)
	}

	if req.GetIndividual() != nil {
		cust = model.Customer{
			CustomerType: int16(commonv1.CustomerType_CUSTOMER_TYPE_INDIVIDUAL),
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			FullName:     &fullName,
			Email:        req.Email,
			KycStatus:    int16(commonv1.KYCStatus_KYC_STATUS_PENDING),
		}
		individual = individualProtoToModel(req.GetIndividual())
		phones = phonesToModel(req.Phones)
		addresses = addressesToModel(req.Addresses)
		return
	}

	// Business customer
	b := req.GetBusiness()
	cust = model.Customer{
		CustomerType: int16(commonv1.CustomerType_CUSTOMER_TYPE_BUSINESS),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		FullName:     &fullName,
		Email:        req.Email,
		KycStatus:    int16(commonv1.KYCStatus_KYC_STATUS_PENDING),
	}
	business = businessProtoToModel(b)
	bizPhones = phonesToModel(b.CompanyPhones)
	bizAddrs = addressesToModel(b.RegisteredAddresses)
	if b.Proprietor != nil {
		propPhones = phonesToModel(b.Proprietor.Phones)
	}
	return
}

// ── private helpers ──────────────────────────────────────────────────────────

func phonesToProto(ms []model.Phone) []*commonv1.Phone {
	out := make([]*commonv1.Phone, len(ms))
	for i, p := range ms {
		out[i] = &commonv1.Phone{
			Type:        commonv1.PhoneType(p.PhoneType),
			CountryCode: p.CountryCode,
			Number:      p.Number,
			IsPrimary:   p.IsPrimary,
		}
	}
	return out
}

func addressesToProto(ms []model.Address) []*commonv1.Address {
	out := make([]*commonv1.Address, len(ms))
	for i, a := range ms {
		out[i] = &commonv1.Address{
			Type:       commonv1.AddressType(a.AddressType),
			Line1:      a.Line1,
			Line2:      derefString(a.Line2),
			City:       a.City,
			State:      a.State,
			PostalCode: a.PostalCode,
			Country:    a.Country,
			IsPrimary:  a.IsPrimary,
		}
	}
	return out
}

func phonesToModel(ps []*commonv1.Phone) []model.Phone {
	out := make([]model.Phone, len(ps))
	for i, p := range ps {
		out[i] = model.Phone{
			PhoneType:   int16(p.Type),
			CountryCode: p.CountryCode,
			Number:      p.Number,
			IsPrimary:   p.IsPrimary,
		}
	}
	return out
}

func addressesToModel(as []*commonv1.Address) []model.Address {
	out := make([]model.Address, len(as))
	for i, a := range as {
		out[i] = model.Address{
			AddressType: int16(a.Type),
			Line1:       a.Line1,
			Line2:       nilIfEmpty(a.Line2),
			City:        a.City,
			State:       a.State,
			PostalCode:  a.PostalCode,
			Country:     a.Country,
			IsPrimary:   a.IsPrimary,
		}
	}
	return out
}

func individualProtoToModel(ind *v1.IndividualDetails) *model.IndividualCustomer {
	m := &model.IndividualCustomer{}
	if ind.NationalId != "" {
		n := ind.NationalId
		m.NationalID = &n
	}
	if ind.Nationality != "" {
		n := ind.Nationality
		m.Nationality = &n
	}
	if ind.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", ind.DateOfBirth)
		if err == nil {
			m.DateOfBirth = &t
		}
	}
	return m
}

func businessProtoToModel(b *v1.BusinessDetails) *model.BusinessCustomer {
	m := &model.BusinessCustomer{
		CompanyName:   b.CompanyName,
		PropFirstName: b.GetProprietor().GetFirstName(),
		PropLastName:  b.GetProprietor().GetLastName(),
		PropEmail:     b.GetProprietor().GetEmail(),
	}
	if b.RegistrationNumber != "" {
		r := b.RegistrationNumber
		m.RegistrationNumber = &r
	}
	if b.TaxId != "" {
		t := b.TaxId
		m.TaxID = &t
	}
	if b.GetProprietor().GetNationalId() != "" {
		n := b.GetProprietor().GetNationalId()
		m.PropNationalID = &n
	}
	return m
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
