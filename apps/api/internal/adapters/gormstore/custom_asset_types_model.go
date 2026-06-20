package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"gorm.io/gorm"
	"time"
)

type customAssetTypeModel struct {
	ID             string          `gorm:"primaryKey;size:26"`
	TenantID       string          `gorm:"not null;size:26;index"`
	Tenant         tenantModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID    *string         `gorm:"size:26;index"`
	Inventory      *inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	Scope          string          `gorm:"not null;size:32;index;check:chk_custom_asset_types_scope,scope IN ('tenant','inventory')"`
	CursorKey      string          `gorm:"not null;size:32;index"`
	TypeKey        string          `gorm:"not null;size:80;index"`
	DisplayName    string          `gorm:"not null;size:120"`
	Description    string          `gorm:"not null;default:'';size:1000"`
	LifecycleState string          `gorm:"not null;default:'active';size:32;index;check:chk_custom_asset_types_lifecycle_state,lifecycle_state IN ('active','archived')"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (customAssetTypeModel) TableName() string {
	return "custom_asset_types"
}

func (m *customAssetTypeModel) BeforeSave(*gorm.DB) error {
	scope := customfield.Scope(m.Scope)
	prefix := "1:"
	if scope == customfield.ScopeTenant {
		prefix = "0:"
	}
	m.CursorKey = prefix + m.ID
	return nil
}

func (m customAssetTypeModel) toDomain() (customfield.AssetType, bool) {
	id, ok := customfield.NewAssetTypeID(m.ID)
	if !ok {
		return customfield.AssetType{}, false
	}
	key, ok := customfield.NewKey(m.TypeKey)
	if !ok {
		return customfield.AssetType{}, false
	}
	displayName, ok := customfield.NewDisplayName(m.DisplayName)
	if !ok {
		return customfield.AssetType{}, false
	}
	description, ok := customfield.NewDescription(m.Description)
	if !ok {
		return customfield.AssetType{}, false
	}
	lifecycleState, ok := customfield.NewAssetTypeLifecycleState(m.LifecycleState)
	if !ok {
		return customfield.AssetType{}, false
	}
	scope := customfield.Scope(m.Scope)
	inventoryID := customfield.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = customfield.InventoryID(*m.InventoryID)
	}
	return customfield.NewAssetTypeWithLifecycle(
		id,
		customfield.TenantID(m.TenantID),
		inventoryID,
		scope,
		key,
		displayName,
		description,
		lifecycleState,
	)
}
