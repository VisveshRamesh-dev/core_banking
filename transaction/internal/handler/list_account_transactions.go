package handler

import (
	"context"
	"fmt"
	"strconv"

	"transaction/internal/data"
	"transaction/internal/mapper"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	transactionv1 "github.com/visvesh-ramesh/corebank/v1/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultPageSize = 20
const maxPageSize = 100

func (h *TransactionHandler) ListAccountTransactions(ctx context.Context, req *transactionv1.ListAccountTransactionsRequest) (*transactionv1.ListAccountTransactionsResponse, error) {
	if req.AccountId == 0 {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
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
			if n, err := strconv.Atoi(req.Page.PageToken); err == nil {
				offset = n
			}
		}
	}

	txs, total, err := h.repo.ListByAccount(ctx, req.AccountId, data.ListParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	proto := make([]*transactionv1.Transaction, len(txs))
	for i, t := range txs {
		proto[i] = mapper.TransactionToProto(t)
	}

	nextToken := ""
	nextOffset := offset + len(txs)
	if int64(nextOffset) < total {
		nextToken = fmt.Sprintf("%d", nextOffset)
	}

	return &transactionv1.ListAccountTransactionsResponse{
		Transactions: proto,
		Page: &commonv1.PageResponse{
			TotalSize:     total,
			NextPageToken: nextToken,
		},
	}, nil
}
