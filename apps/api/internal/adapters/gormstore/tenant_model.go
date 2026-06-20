package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"time"
)

type tenantModel struct {
	ID             string `gorm:"primaryKey;size:26"`
	Name           string `gorm:"not null;size:120"`
	LifecycleState string `gorm:"not null;size:32;default:'active'"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (tenantModel) TableName() string {
	return "tenants"
}

func (m tenantModel) toDomain() (tenant.Tenant, bool) {
	id, ok := tenant.NewID(m.ID)
	if !ok {
		return tenant.Tenant{}, false
	}
	name, ok := tenant.NewName(m.Name)
	if !ok {
		return tenant.Tenant{}, false
	}
	lifecycleState, ok := tenant.NewLifecycleState(m.LifecycleState)
	if !ok {
		return tenant.Tenant{}, false
	}
	return tenant.NewTenant(id, name, lifecycleState)
}
