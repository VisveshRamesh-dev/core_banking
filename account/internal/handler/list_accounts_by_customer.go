package handler

import (
	"context"

	"account/internal/data"
	"account/internal/mapper"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultPageSize = 20
const maxPageSize = 100

func (h *AccountHandler) ListAccountsByCustomer(ctx context.Context, req *accountv1.ListAccountsByCustomerRequest) (*accountv1.ListAccountsByCustomerResponse, error) {
	if req.CustomerId == 0 {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	limit := defaultPageSize
	offset := 0
	if req.Page != nil {
		if req.Page.PageSize > 0 {
			limit = int(req.Page.PageSize)
			if limit > maxPageSize {
				limit = maxPageSize
			}
		}
		if req.Page.PageToken != "" {
			// PageToken is an opaque offset string — parse as integer offset.
			var off int
			if _, err := parseOffset(req.Page.PageToken, &off); err == nil {
				offset = off
			}
		}
	}

	accounts, total, err := h.repo.ListByCustomer(ctx, req.CustomerId, data.ListParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := make([]*accountv1.Account, len(accounts))
	for i, a := range accounts {
		proto[i] = mapper.AccountToProto(a)
	}

	nextToken := ""
	nextOffset := offset + len(accounts)
	if int64(nextOffset) < total {
		nextToken = formatOffset(nextOffset)
	}

	return &accountv1.ListAccountsByCustomerResponse{
		Accounts: proto,
		Page: &commonv1.PageResponse{
			TotalSize:     total,
			NextPageToken: nextToken,
		},
	}, nil
}
