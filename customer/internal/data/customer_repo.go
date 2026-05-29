package data

import (
	"context"
	"errors"

	"customer/internal/model"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// CustomerRecord bundles a customer with all related rows so the handler
// layer never needs to issue multiple repo calls.
type CustomerRecord struct {
	Customer   model.Customer
	Individual *model.IndividualCustomer
	Business   *model.BusinessCustomer
	// Individual contacts (link_type = LinkTypeIndividual)
	Phones    []model.Phone
	Addresses []model.Address
	// Business company contacts (link_type = LinkTypeBusiness)
	BizPhones []model.Phone
	BizAddrs  []model.Address
	// Proprietor contacts (link_type = LinkTypeProprietor)
	PropPhones []model.Phone
}

// CustomerRepo is the only data-access object for the customer aggregate.
type CustomerRepo struct {
	db *gorm.DB
}

func NewCustomerRepo(data *Data) *CustomerRepo {
	return &CustomerRepo{db: data.db}
}

// Create persists a full customer aggregate in a single transaction.
// All generated IDs (individual_id, business_id) are back-filled into rec
// before the function returns.
func (r *CustomerRepo) Create(ctx context.Context, rec *CustomerRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if rec.Individual != nil {
			if err := tx.Create(rec.Individual).Error; err != nil {
				return err
			}
			id := rec.Individual.ID
			rec.Customer.IndividualID = &id

			if err := r.insertContacts(tx, rec.Individual.ID, model.LinkTypeIndividual, rec.Phones, rec.Addresses); err != nil {
				return err
			}
		}

		if rec.Business != nil {
			if err := tx.Create(rec.Business).Error; err != nil {
				return err
			}
			bid := rec.Business.ID
			rec.Customer.BusinessID = &bid

			if err := r.insertContacts(tx, bid, model.LinkTypeBusiness, rec.BizPhones, rec.BizAddrs); err != nil {
				return err
			}
			if err := r.insertContacts(tx, bid, model.LinkTypeProprietor, rec.PropPhones, nil); err != nil {
				return err
			}
		}

		return tx.Create(&rec.Customer).Error
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

// UpdateKYCStatus sets kyc_status and returns the refreshed full record.
func (r *CustomerRepo) UpdateKYCStatus(ctx context.Context, id int64, newStatus int16) (*CustomerRecord, error) {
	res := r.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ?", id).
		Update("kyc_status", newStatus)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
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

// ── private helpers ───────────────────────────────────────────────────────────

// insertContacts inserts phone and address rows, then creates rel_contact links.
func (r *CustomerRepo) insertContacts(
	tx *gorm.DB,
	linkID int64,
	linkType int16,
	phones []model.Phone,
	addresses []model.Address,
) error {
	for i := range phones {
		if err := tx.Create(&phones[i]).Error; err != nil {
			return err
		}
		link := model.RelContact{
			ContactID:   phones[i].ID,
			ContactType: model.ContactTypePhone,
			LinkID:      linkID,
			LinkType:    linkType,
		}
		if err := tx.Create(&link).Error; err != nil {
			return err
		}
	}
	for i := range addresses {
		if err := tx.Create(&addresses[i]).Error; err != nil {
			return err
		}
		link := model.RelContact{
			ContactID:   addresses[i].ID,
			ContactType: model.ContactTypeAddress,
			LinkID:      linkID,
			LinkType:    linkType,
		}
		if err := tx.Create(&link).Error; err != nil {
			return err
		}
	}
	return nil
}

// loadRelated fetches all related rows for a customer.
func (r *CustomerRepo) loadRelated(ctx context.Context, c model.Customer) (*CustomerRecord, error) {
	rec := &CustomerRecord{Customer: c}

	if c.IndividualID != nil {
		var ind model.IndividualCustomer
		if err := r.db.WithContext(ctx).First(&ind, *c.IndividualID).Error; err != nil {
			return nil, err
		}
		rec.Individual = &ind

		phones, addrs, err := r.loadContacts(ctx, *c.IndividualID, model.LinkTypeIndividual)
		if err != nil {
			return nil, err
		}
		rec.Phones = phones
		rec.Addresses = addrs
	}

	if c.BusinessID != nil {
		var biz model.BusinessCustomer
		if err := r.db.WithContext(ctx).First(&biz, *c.BusinessID).Error; err != nil {
			return nil, err
		}
		rec.Business = &biz

		bizPhones, bizAddrs, err := r.loadContacts(ctx, *c.BusinessID, model.LinkTypeBusiness)
		if err != nil {
			return nil, err
		}
		rec.BizPhones = bizPhones
		rec.BizAddrs = bizAddrs

		propPhones, _, err := r.loadContacts(ctx, *c.BusinessID, model.LinkTypeProprietor)
		if err != nil {
			return nil, err
		}
		rec.PropPhones = propPhones
	}

	return rec, nil
}

// loadContacts queries rel_contact for a given owner, then fetches each phone/address.
func (r *CustomerRepo) loadContacts(ctx context.Context, linkID int64, linkType int16) ([]model.Phone, []model.Address, error) {
	var links []model.RelContact
	if err := r.db.WithContext(ctx).
		Where("link_id = ? AND link_type = ?", linkID, linkType).
		Find(&links).Error; err != nil {
		return nil, nil, err
	}

	var phoneIDs, addrIDs []int64
	for _, l := range links {
		switch l.ContactType {
		case model.ContactTypePhone:
			phoneIDs = append(phoneIDs, l.ContactID)
		case model.ContactTypeAddress:
			addrIDs = append(addrIDs, l.ContactID)
		}
	}

	var phones []model.Phone
	if len(phoneIDs) > 0 {
		if err := r.db.WithContext(ctx).Where("id IN ?", phoneIDs).Find(&phones).Error; err != nil {
			return nil, nil, err
		}
	}

	var addresses []model.Address
	if len(addrIDs) > 0 {
		if err := r.db.WithContext(ctx).Where("id IN ?", addrIDs).Find(&addresses).Error; err != nil {
			return nil, nil, err
		}
	}

	return phones, addresses, nil
}
