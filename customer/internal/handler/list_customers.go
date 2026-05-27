package handler

import (
	"context"
	"fmt"

	"customer/internal/data"
	"customer/internal/mapper"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	v1 "github.com/visvesh-ramesh/corebank/v1/customer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultPageSize = 20

func (h *CustomerHandler) ListCustomers(ctx context.Context, req *v1.ListCustomersRequest) (*v1.ListCustomersResponse, error) {
	pageSize := int(req.GetPage().GetPageSize())
	if pageSize <= 0 || pageSize > 100 {
		pageSize = defaultPageSize
	}

	offset := 0
	if tok := req.GetPage().GetPageToken(); tok != "" {
		fmt.Sscanf(tok, "%d", &offset) //nolint:errcheck
	}

	records, total, err := h.customerRepo.List(ctx, data.ListParams{
		StatusFilter: int16(req.StatusFilter),
		TypeFilter:   int16(req.TypeFilter),
		Limit:        pageSize,
		Offset:       offset,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	customers := make([]*v1.Customer, len(records))
	for i, rec := range records {
		customers[i] = mapper.CustomerToProto(
			rec.Customer, rec.Phones, rec.Addresses,
			rec.Individual, rec.Business,
			rec.BizPhones, rec.BizAddrs, rec.Proprietor, rec.PropPhones,
		)
	}

	nextToken := ""
	if nextOffset := offset + len(records); int64(nextOffset) < total {
		nextToken = fmt.Sprintf("%d", nextOffset)
	}

	return &v1.ListCustomersResponse{
		Customers: customers,
		Page: &commonv1.PageResponse{
			NextPageToken: nextToken,
			TotalSize:     total,
		},
	}, nil
}
