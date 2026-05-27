package data

import (
	"context"
	"errors"

	"customer/internal/model"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// CustomerRecord bundles a customer with all of its related rows so the
// handler layer never needs to issue multiple repo calls.
type CustomerRecord struct {
	Customer   model.Customer
	Phones     []model.CustomerPhone
	Addresses  []model.CustomerAddress
	Individual *model.CustomerIndividualDetail
	Business   *model.CustomerBusinessDetail
	BizPhones  []model.BusinessPhone
	BizAddrs   []model.BusinessAddress
	Proprietor *model.BusinessProprietor
	PropPhones []model.ProprietorPhone
}

// CustomerRepo is the only data-access object for the customer aggregate.
type CustomerRepo struct {
	db *gorm.DB
}

func NewCustomerRepo(data *Data) *CustomerRepo {
	return &CustomerRepo{db: data.db}
}

// Create persists a full customer aggregate in a single transaction.
// All IDs (customer_id, business_id, proprietor_id) inside rec are
// back-filled from database-generated values before the function returns.
func (r *CustomerRepo) Create(ctx context.Context, rec *CustomerRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rec.Customer).Error; err != nil {
			return err
		}
		cid := rec.Customer.ID

		for i := range rec.Phones {
			rec.Phones[i].CustomerID = cid
		}
		if len(rec.Phones) > 0 {
			if err := tx.Create(&rec.Phones).Error; err != nil {
				return err
			}
		}

		for i := range rec.Addresses {
			rec.Addresses[i].CustomerID = cid
		}
		if len(rec.Addresses) > 0 {
			if err := tx.Create(&rec.Addresses).Error; err != nil {
				return err
			}
		}

		if rec.Individual != nil {
			rec.Individual.CustomerID = cid
			if err := tx.Create(rec.Individual).Error; err != nil {
				return err
			}
		}

		if rec.Business != nil {
			rec.Business.CustomerID = cid
			if err := tx.Create(rec.Business).Error; err != nil {
				return err
			}
			bid := rec.Business.ID

			for i := range rec.BizPhones {
				rec.BizPhones[i].BusinessID = bid
			}
			if len(rec.BizPhones) > 0 {
				if err := tx.Create(&rec.BizPhones).Error; err != nil {
					return err
				}
			}

			for i := range rec.BizAddrs {
				rec.BizAddrs[i].BusinessID = bid
			}
			if len(rec.BizAddrs) > 0 {
				if err := tx.Create(&rec.BizAddrs).Error; err != nil {
					return err
				}
			}

			if rec.Proprietor != nil {
				rec.Proprietor.BusinessID = bid
				if err := tx.Create(rec.Proprietor).Error; err != nil {
					return err
				}
				pid := rec.Proprietor.ID

				for i := range rec.PropPhones {
					rec.PropPhones[i].ProprietorID = pid
				}
				if len(rec.PropPhones) > 0 {
					if err := tx.Create(&rec.PropPhones).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

// GetByID fetches a customer and all related rows by primary key.
func (r *CustomerRepo) GetByID(ctx context.Context, id int64) (*CustomerRecord, error) {
	var c model.Customer
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.loadRelated(ctx, c)
}

// UpdateKYCStatus sets kyc_status on the customer row and returns the full record.
func (r *CustomerRepo) UpdateKYCStatus(ctx context.Context, id int64, newStatus int16) (*CustomerRecord, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ?", id).
		Update("kyc_status", newStatus)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// ListParams carries optional filters and pagination for List.
type ListParams struct {
	StatusFilter int16
	TypeFilter   int16
	Limit        int
	Offset       int
}

// List returns paginated customers matching the given filters plus a total count.
func (r *CustomerRepo) List(ctx context.Context, p ListParams) ([]*CustomerRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Customer{})
	if p.StatusFilter != 0 {
		q = q.Where("kyc_status = ?", p.StatusFilter)
	}
	if p.TypeFilter != 0 {
		q = q.Where("customer_type = ?", p.TypeFilter)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var customers []model.Customer
	if err := q.Order("id").Limit(p.Limit).Offset(p.Offset).Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	records := make([]*CustomerRecord, len(customers))
	for i, c := range customers {
		rec, err := r.loadRelated(ctx, c)
		if err != nil {
			return nil, 0, err
		}
		records[i] = rec
	}
	return records, total, nil
}

// loadRelated fetches all related rows for a customer record.
func (r *CustomerRepo) loadRelated(ctx context.Context, c model.Customer) (*CustomerRecord, error) {
	rec := &CustomerRecord{Customer: c}
	cid := c.ID

	if err := r.db.WithContext(ctx).Where("customer_id = ?", cid).Find(&rec.Phones).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Where("customer_id = ?", cid).Find(&rec.Addresses).Error; err != nil {
		return nil, err
	}

	var ind model.CustomerIndividualDetail
	if err := r.db.WithContext(ctx).Where("customer_id = ?", cid).First(&ind).Error; err == nil {
		rec.Individual = &ind
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var biz model.CustomerBusinessDetail
	if err := r.db.WithContext(ctx).Where("customer_id = ?", cid).First(&biz).Error; err == nil {
		rec.Business = &biz
		bid := biz.ID

		if err := r.db.WithContext(ctx).Where("business_id = ?", bid).Find(&rec.BizPhones).Error; err != nil {
			return nil, err
		}
		if err := r.db.WithContext(ctx).Where("business_id = ?", bid).Find(&rec.BizAddrs).Error; err != nil {
			return nil, err
		}

		var prop model.BusinessProprietor
		if err := r.db.WithContext(ctx).Where("business_id = ?", bid).First(&prop).Error; err == nil {
			rec.Proprietor = &prop
			if err := r.db.WithContext(ctx).Where("proprietor_id = ?", prop.ID).Find(&rec.PropPhones).Error; err != nil {
				return nil, err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return rec, nil
}
