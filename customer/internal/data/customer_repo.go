package data

import "gorm.io/gorm"

type CustomerRepo struct {
	db *gorm.DB
}

func NewCustomerRepo(data *Data) *CustomerRepo {
	return &CustomerRepo{db: data.db}
}
func (r *CustomerRepo) Find() {
}
