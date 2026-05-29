package model

import (
	"utils/snowflake"

	"gorm.io/gorm"
)

func (m *Customer) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}

func (m *IndividualCustomer) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}

func (m *BusinessCustomer) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}

func (m *Phone) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}

func (m *Address) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}

func (m *RelContact) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}
