package mapper

import (
	"account/internal/model"

	commonv1 "github.com/visvesh-ramesh/corebank/v1/common"
	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AccountToProto(a *model.Account) *accountv1.Account {
	return &accountv1.Account{
		Id:         a.ID,
		CustomerId: a.CustomerID,
		Type:       commonv1.AccountType(a.Type),
		Status:     commonv1.AccountStatus(a.Status),
		Currency:   a.Currency,
		CachedBalance: &commonv1.Money{
			AmountMinor: a.CachedBalanceMinor,
			Currency:    a.Currency,
		},
		OverdraftLimitMinor: a.OverdraftLimitMinor,
		CreatedAt:           timestamppb.New(a.CreatedAt),
		UpdatedAt:           timestamppb.New(a.UpdatedAt),
	}
}
