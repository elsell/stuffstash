package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"time"
)

type assetModel struct {
	ID                string                `gorm:"primaryKey;size:26"`
	TenantID          string                `gorm:"not null;size:26;index:idx_assets_tenant_inventory"`
	Tenant            tenantModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID       string                `gorm:"not null;size:26;index:idx_assets_tenant_inventory;index:idx_assets_inventory_parent;index:idx_assets_inventory_kind"`
	Inventory         inventoryModel        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	ParentAssetID     *string               `gorm:"size:26;index;index:idx_assets_inventory_parent"`
	ParentAsset       *assetModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:ParentAssetID;references:ID"`
	CustomAssetTypeID *string               `gorm:"size:26;index"`
	CustomAssetType   *customAssetTypeModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:CustomAssetTypeID;references:ID"`
	Kind              string                `gorm:"not null;size:32;index:idx_assets_inventory_kind;check:chk_assets_kind,kind IN ('item','container','location')"`
	Title             string                `gorm:"not null;size:160"`
	Description       string                `gorm:"not null;default:''"`
	CustomFields      string                `gorm:"type:jsonb;not null;default:'{}'"`
	LifecycleState    string                `gorm:"not null;size:32;check:chk_assets_lifecycle_state,lifecycle_state IN ('active','archived')"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (assetModel) TableName() string {
	return "assets"
}

func (m assetModel) toDomain() (asset.Asset, bool) {
	id, ok := asset.NewID(m.ID)
	if !ok {
		return asset.Asset{}, false
	}
	kind, ok := asset.NewKind(m.Kind)
	if !ok {
		return asset.Asset{}, false
	}
	title, ok := asset.NewTitle(m.Title)
	if !ok {
		return asset.Asset{}, false
	}
	var customFieldValues map[string]any
	if err := json.Unmarshal([]byte(m.CustomFields), &customFieldValues); err != nil {
		return asset.Asset{}, false
	}
	customFields, ok := asset.NewCustomFields(customFieldValues)
	if !ok {
		return asset.Asset{}, false
	}
	lifecycleState := asset.LifecycleState(m.LifecycleState)
	switch lifecycleState {
	case asset.LifecycleStateActive, asset.LifecycleStateArchived:
	default:
		return asset.Asset{}, false
	}
	parentID := asset.ID("")
	if m.ParentAssetID != nil {
		parentID, ok = asset.NewID(*m.ParentAssetID)
		if !ok {
			return asset.Asset{}, false
		}
	}
	customAssetTypeID := asset.CustomAssetTypeID("")
	if m.CustomAssetTypeID != nil {
		customAssetTypeID, ok = asset.NewCustomAssetTypeID(*m.CustomAssetTypeID)
		if !ok {
			return asset.Asset{}, false
		}
	}
	return asset.Asset{
		ID:                id,
		TenantID:          asset.TenantID(m.TenantID),
		InventoryID:       asset.InventoryID(m.InventoryID),
		ParentAssetID:     parentID,
		CustomAssetTypeID: customAssetTypeID,
		Kind:              kind,
		Title:             title,
		Description:       asset.NewDescription(m.Description),
		CustomFields:      customFields,
		LifecycleState:    lifecycleState,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}, true
}
