package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"time"
)

type inventoryModel struct {
	ID        string      `gorm:"primaryKey;size:26"`
	TenantID  string      `gorm:"not null;size:26;index"`
	Tenant    tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	Name      string      `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (inventoryModel) TableName() string {
	return "inventories"
}

func (m inventoryModel) toDomain() (inventory.Inventory, bool) {
	id, ok := inventory.NewID(m.ID)
	if !ok {
		return inventory.Inventory{}, false
	}
	name, ok := inventory.NewName(m.Name)
	if !ok {
		return inventory.Inventory{}, false
	}

	return inventory.Inventory{
		ID:       id,
		TenantID: inventory.TenantID(m.TenantID),
		Name:     name,
	}, true
}
