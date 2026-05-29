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
	if req.GetIndividual() == nil && req.GetBusiness() == nil {
		return nil, status.Error(codes.InvalidArgument, "exactly one of individual or business details must be set")
	}
	if req.GetIndividual() != nil && (len(req.Phones) == 0 || len(req.Addresses) == 0) {
		return nil, status.Error(codes.InvalidArgument, "individual customers require at least one phone and one address")
	}
	if b := req.GetBusiness(); b != nil {
		if len(b.CompanyPhones) == 0 || len(b.RegisteredAddresses) == 0 {
			return nil, status.Error(codes.InvalidArgument, "business customers require at least one company phone and one registered address")
		}
		if b.Proprietor == nil {
			return nil, status.Error(codes.InvalidArgument, "business customers require proprietor information")
		}
	}

	cust, individual, business, phones, addresses, bizPhones, bizAddrs, propPhones :=
		mapper.OnboardRequestToModels(req)

	rec := &data.CustomerRecord{
		Customer:   cust,
		Individual: individual,
		Business:   business,
		Phones:     phones,
		Addresses:  addresses,
		BizPhones:  bizPhones,
		BizAddrs:   bizAddrs,
		PropPhones: propPhones,
	}

	if err := h.customerRepo.Create(ctx, rec); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := mapper.CustomerToProto(
		rec.Customer,
		rec.Individual, rec.Business,
		rec.Phones, rec.Addresses,
		rec.BizPhones, rec.BizAddrs,
		rec.PropPhones,
	)
	return &v1.OnboardResponse{Customer: proto}, nil
}
