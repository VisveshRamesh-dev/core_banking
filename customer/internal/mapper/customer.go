package mapper

import (
	"fmt"
	"time"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/protobuf/types/known/timestamppb"

	"customer/internal/model"
)

// CustomerToProto assembles a full v1.Customer proto from the database aggregate.
func CustomerToProto(
	c model.Customer,
	phones []model.CustomerPhone,
	addresses []model.CustomerAddress,
	individual *model.CustomerIndividualDetail,
	business *model.CustomerBusinessDetail,
	bizPhones []model.BusinessPhone,
	bizAddrs []model.BusinessAddress,
	proprietor *model.BusinessProprietor,
	propPhones []model.ProprietorPhone,
) *v1.Customer {
	protoPhones := make([]*commonv1.Phone, len(phones))
	for i, p := range phones {
		protoPhones[i] = customerPhoneToProto(p)
	}

	protoAddrs := make([]*commonv1.Address, len(addresses))
	for i, a := range addresses {
		protoAddrs[i] = customerAddressToProto(a)
	}

	cust := &v1.Customer{
		Id:           c.ID,
		CustomerType: commonv1.CustomerType(c.CustomerType),
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		FullName:     derefString(c.FullName),
		Email:        c.Email,
		Phones:       protoPhones,
		Addresses:    protoAddrs,
		KycStatus:    commonv1.KYCStatus(c.KycStatus),
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
	}

	switch {
	case individual != nil:
		ind := &v1.IndividualDetails{
			NationalId:  derefString(individual.NationalID),
			Nationality: derefString(individual.Nationality),
		}
		if individual.DateOfBirth != nil {
			ind.DateOfBirth = individual.DateOfBirth.Format("2006-01-02")
		}
		cust.Details = &v1.Customer_Individual{Individual: ind}

	case business != nil:
		biz := &v1.BusinessDetails{
			CompanyName:        business.CompanyName,
			RegistrationNumber: derefString(business.RegistrationNumber),
			TaxId:              derefString(business.TaxID),
		}

		biz.CompanyPhones = make([]*commonv1.Phone, len(bizPhones))
		for i, p := range bizPhones {
			biz.CompanyPhones[i] = businessPhoneToProto(p)
		}

		biz.RegisteredAddresses = make([]*commonv1.Address, len(bizAddrs))
		for i, a := range bizAddrs {
			biz.RegisteredAddresses[i] = businessAddressToProto(a)
		}

		if proprietor != nil {
			prop := &v1.ProprietorInfo{
				FirstName:  proprietor.FirstName,
				LastName:   proprietor.LastName,
				Email:      proprietor.Email,
				NationalId: derefString(proprietor.NationalID),
			}
			prop.Phones = make([]*commonv1.Phone, len(propPhones))
			for i, p := range propPhones {
				prop.Phones[i] = proprietorPhoneToProto(p)
			}
			biz.Proprietor = prop
		}
		cust.Details = &v1.Customer_Business{Business: biz}
	}

	return cust
}

// OnboardRequestToCustomer builds the root Customer DB row from the request.
// The ID is left zero — GORM fills it after INSERT.
func OnboardRequestToCustomer(req *v1.OnboardRequest) model.Customer {
	customerType := int16(commonv1.CustomerType_CUSTOMER_TYPE_INDIVIDUAL)
	if req.GetBusiness() != nil {
		customerType = int16(commonv1.CustomerType_CUSTOMER_TYPE_BUSINESS)
	}

	fullName := req.FullName
	if fullName == "" {
		fullName = fmt.Sprintf("%s %s", req.FirstName, req.LastName)
	}

	return model.Customer{
		CustomerType: customerType,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		FullName:     &fullName,
		Email:        req.Email,
		KycStatus:    int16(commonv1.KYCStatus_KYC_STATUS_PENDING),
	}
}

// PhonesToCustomerModels converts proto Phone slices to CustomerPhone rows.
// customerID is set by the caller after the customer row is inserted.
func PhonesToCustomerModels(phones []*commonv1.Phone, customerID int64) []model.CustomerPhone {
	out := make([]model.CustomerPhone, len(phones))
	for i, p := range phones {
		out[i] = model.CustomerPhone{
			CustomerID:  customerID,
			PhoneType:   int16(p.Type),
			CountryCode: p.CountryCode,
			Number:      p.Number,
			IsPrimary:   p.IsPrimary,
		}
	}
	return out
}

// AddressesToCustomerModels converts proto Address slices to CustomerAddress rows.
func AddressesToCustomerModels(addresses []*commonv1.Address, customerID int64) []model.CustomerAddress {
	out := make([]model.CustomerAddress, len(addresses))
	for i, a := range addresses {
		out[i] = model.CustomerAddress{
			CustomerID:  customerID,
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

// IndividualToModel converts IndividualDetails proto into a DB row.
func IndividualToModel(ind *v1.IndividualDetails, customerID int64) model.CustomerIndividualDetail {
	detail := model.CustomerIndividualDetail{CustomerID: customerID}
	if ind.NationalId != "" {
		n := ind.NationalId
		detail.NationalID = &n
	}
	if ind.Nationality != "" {
		n := ind.Nationality
		detail.Nationality = &n
	}
	if ind.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", ind.DateOfBirth)
		if err == nil {
			detail.DateOfBirth = &t
		}
	}
	return detail
}

// BusinessToModels converts BusinessDetails proto into all related DB rows.
// Foreign-key IDs (business_id, proprietor_id) are left zero and must be
// filled in by the caller after the parent row is inserted.
func BusinessToModels(biz *v1.BusinessDetails, customerID int64) (
	detail model.CustomerBusinessDetail,
	bizPhones []model.BusinessPhone,
	bizAddrs []model.BusinessAddress,
	proprietor model.BusinessProprietor,
	propPhones []model.ProprietorPhone,
) {
	detail = model.CustomerBusinessDetail{
		CustomerID:  customerID,
		CompanyName: biz.CompanyName,
	}
	if biz.RegistrationNumber != "" {
		r := biz.RegistrationNumber
		detail.RegistrationNumber = &r
	}
	if biz.TaxId != "" {
		t := biz.TaxId
		detail.TaxID = &t
	}

	bizPhones = make([]model.BusinessPhone, len(biz.CompanyPhones))
	for i, p := range biz.CompanyPhones {
		bizPhones[i] = model.BusinessPhone{
			PhoneType:   int16(p.Type),
			CountryCode: p.CountryCode,
			Number:      p.Number,
			IsPrimary:   p.IsPrimary,
		}
	}

	bizAddrs = make([]model.BusinessAddress, len(biz.RegisteredAddresses))
	for i, a := range biz.RegisteredAddresses {
		bizAddrs[i] = model.BusinessAddress{
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

	if biz.Proprietor != nil {
		p := biz.Proprietor
		proprietor = model.BusinessProprietor{
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Email:     p.Email,
		}
		if p.NationalId != "" {
			n := p.NationalId
			proprietor.NationalID = &n
		}
		propPhones = make([]model.ProprietorPhone, len(p.Phones))
		for i, ph := range p.Phones {
			propPhones[i] = model.ProprietorPhone{
				PhoneType:   int16(ph.Type),
				CountryCode: ph.CountryCode,
				Number:      ph.Number,
				IsPrimary:   ph.IsPrimary,
			}
		}
	}

	return
}

// ── helpers ──────────────────────────────────────────────────────────────────

func customerPhoneToProto(p model.CustomerPhone) *commonv1.Phone {
	return &commonv1.Phone{
		Type:        commonv1.PhoneType(p.PhoneType),
		CountryCode: p.CountryCode,
		Number:      p.Number,
		IsPrimary:   p.IsPrimary,
	}
}

func customerAddressToProto(a model.CustomerAddress) *commonv1.Address {
	return &commonv1.Address{
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

func businessPhoneToProto(p model.BusinessPhone) *commonv1.Phone {
	return &commonv1.Phone{
		Type:        commonv1.PhoneType(p.PhoneType),
		CountryCode: p.CountryCode,
		Number:      p.Number,
		IsPrimary:   p.IsPrimary,
	}
}

func businessAddressToProto(a model.BusinessAddress) *commonv1.Address {
	return &commonv1.Address{
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

func proprietorPhoneToProto(p model.ProprietorPhone) *commonv1.Phone {
	return &commonv1.Phone{
		Type:        commonv1.PhoneType(p.PhoneType),
		CountryCode: p.CountryCode,
		Number:      p.Number,
		IsPrimary:   p.IsPrimary,
	}
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
