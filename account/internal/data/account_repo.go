package data

import (
	"context"
	"errors"

	"account/internal/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

// ListParams carries pagination for listing accounts by customer.
type ListParams struct {
	Limit  int
	Offset int
}

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(data *Data) *AccountRepo {
	return &AccountRepo{db: data.db}
}

func (r *AccountRepo) Create(ctx context.Context, a *model.Account) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AccountRepo) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	var a model.Account
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// UpdateStatus sets the account status and returns the refreshed record.
func (r *AccountRepo) UpdateStatus(ctx context.Context, id int64, newStatus int16) (*model.Account, error) {
	res := r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("id = ?", id).
		Update("status", newStatus)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// ListByCustomer returns paginated accounts for a customer plus total count.
func (r *AccountRepo) ListByCustomer(ctx context.Context, customerID int64, p ListParams) ([]*model.Account, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Account{}).Where("customer_id = ?", customerID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var accounts []model.Account
	if err := q.Order("id").Limit(p.Limit).Offset(p.Offset).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*model.Account, len(accounts))
	for i := range accounts {
		result[i] = &accounts[i]
	}
	return result, total, nil
}

// UpdateCachedBalance updates the cached balance projection on an account.
func (r *AccountRepo) UpdateCachedBalance(ctx context.Context, id int64, balanceMinor int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("id = ?", id).
		Update("cached_balance_minor", balanceMinor).Error
}
