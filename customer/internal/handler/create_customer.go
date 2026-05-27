package handler

import (
	"context"

	"customer/internal/data"
	"customer/internal/mapper"

	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CustomerHandler) Onboard(ctx context.Context, req *v1.OnboardRequest) (*v1.OnboardResponse, error) {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "first_name, last_name, and email are required")
	}
	if len(req.Phones) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one phone is required")
	}
	if len(req.Addresses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one address is required")
	}
	if req.GetIndividual() == nil && req.GetBusiness() == nil {
		return nil, status.Error(codes.InvalidArgument, "exactly one of individual or business details must be set")
	}

	cust := mapper.OnboardRequestToCustomer(req)
	phones := mapper.PhonesToCustomerModels(req.Phones, 0)
	addrs := mapper.AddressesToCustomerModels(req.Addresses, 0)

	rec := &data.CustomerRecord{
		Customer:  cust,
		Phones:    phones,
		Addresses: addrs,
	}

	if req.GetIndividual() != nil {
		ind := mapper.IndividualToModel(req.GetIndividual(), 0)
		rec.Individual = &ind
	} else {
		detail, bizPhones, bizAddrs, prop, propPhones := mapper.BusinessToModels(req.GetBusiness(), 0)
		rec.Business = &detail
		rec.BizPhones = bizPhones
		rec.BizAddrs = bizAddrs
		rec.Proprietor = &prop
		rec.PropPhones = propPhones
	}

	if err := h.customerRepo.Create(ctx, rec); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.CustomerToProto(
		rec.Customer, rec.Phones, rec.Addresses,
		rec.Individual, rec.Business,
		rec.BizPhones, rec.BizAddrs, rec.Proprietor, rec.PropPhones,
	)
	return &v1.OnboardResponse{Customer: proto}, nil
}
