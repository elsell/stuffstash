package gormstore

import (
	"time"
)

type tenantModel struct {
	ID        string `gorm:"primaryKey;size:26"`
	Name      string `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tenantModel) TableName() string {
	return "tenants"
}
