package model

import (
	"utils/snowflake"

	"gorm.io/gorm"
)

func (m *LedgerTransaction) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}

func (m *LedgerEntry) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = snowflake.NextID()
	}
	return nil
}
