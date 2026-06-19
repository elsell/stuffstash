package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"time"
)

type inventoryAccessGrantModel struct {
	TenantID     string         `gorm:"primaryKey;size:26;index:idx_inventory_access_grants_inventory"`
	Tenant       tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID  string         `gorm:"primaryKey;size:26;index:idx_inventory_access_grants_inventory"`
	Inventory    inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	GrantKey     string         `gorm:"primaryKey;size:180"`
	PrincipalID  string         `gorm:"not null;size:128;index"`
	Relationship string         `gorm:"not null;size:32;check:chk_inventory_access_grants_relationship,relationship IN ('viewer','editor')"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (inventoryAccessGrantModel) TableName() string {
	return "inventory_access_grants"
}

func (m *inventoryAccessGrantModel) BeforeSave(*gorm.DB) error {
	m.GrantKey = ports.InventoryAccessGrant{
		PrincipalID:  identity.PrincipalID(m.PrincipalID),
		Relationship: ports.InventoryAccessRelationship(m.Relationship),
	}.CursorKey()
	return nil
}

func (m inventoryAccessGrantModel) toPort() (ports.InventoryAccessGrant, bool) {
	principalID, ok := identity.NewPrincipalID(m.PrincipalID)
	if !ok {
		return ports.InventoryAccessGrant{}, false
	}
	relationship := ports.InventoryAccessRelationship(m.Relationship)
	switch relationship {
	case ports.InventoryAccessViewer, ports.InventoryAccessEditor:
	default:
		return ports.InventoryAccessGrant{}, false
	}
	return ports.InventoryAccessGrant{
		TenantID:     tenant.ID(m.TenantID),
		InventoryID:  inventory.InventoryID(m.InventoryID),
		PrincipalID:  principalID,
		Relationship: relationship,
	}, true
}
