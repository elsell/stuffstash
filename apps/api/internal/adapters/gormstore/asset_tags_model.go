package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
)

type assetTagModel struct {
	ID             string         `gorm:"primaryKey;size:26"`
	TenantID       string         `gorm:"not null;size:26;index:idx_asset_tags_scope_key,unique;index:idx_asset_tags_scope_id"`
	Tenant         tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID    string         `gorm:"not null;size:26;index:idx_asset_tags_scope_key,unique;index:idx_asset_tags_scope_id"`
	Inventory      inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	Key            string         `gorm:"not null;size:80;index:idx_asset_tags_scope_key,unique"`
	DisplayName    string         `gorm:"not null;size:80"`
	Color          string         `gorm:"not null;default:'';size:7"`
	LifecycleState string         `gorm:"not null;size:32;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type assetTagAssignmentModel struct {
	TenantID    string        `gorm:"primaryKey;size:26;index:idx_asset_tag_assignments_asset"`
	InventoryID string        `gorm:"primaryKey;size:26;index:idx_asset_tag_assignments_asset"`
	AssetID     string        `gorm:"primaryKey;size:26;index:idx_asset_tag_assignments_asset"`
	Asset       assetModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:AssetID;references:ID"`
	TagID       string        `gorm:"primaryKey;size:26;index"`
	Tag         assetTagModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:TagID;references:ID"`
	CreatedAt   time.Time
}

func (assetTagAssignmentModel) TableName() string {
	return "asset_tag_assignments"
}

func (assetTagModel) TableName() string {
	return "asset_tags"
}

func assetTagModelFromDomain(tag assettag.Tag) assetTagModel {
	return assetTagModel{
		ID:             tag.ID.String(),
		TenantID:       tag.TenantID.String(),
		InventoryID:    tag.InventoryID.String(),
		Key:            tag.Key.String(),
		DisplayName:    tag.DisplayName.String(),
		Color:          tag.Color.String(),
		LifecycleState: tag.LifecycleState.String(),
		CreatedAt:      tag.CreatedAt,
		UpdatedAt:      tag.UpdatedAt,
	}
}

func (m assetTagModel) toDomain() (assettag.Tag, bool) {
	id, ok := assettag.NewID(m.ID)
	if !ok {
		return assettag.Tag{}, false
	}
	key, ok := assettag.NewKey(m.Key)
	if !ok {
		return assettag.Tag{}, false
	}
	displayName, ok := assettag.NewDisplayName(m.DisplayName)
	if !ok {
		return assettag.Tag{}, false
	}
	color, ok := assettag.NewColor(m.Color)
	if !ok {
		return assettag.Tag{}, false
	}
	state := assettag.LifecycleState(m.LifecycleState)
	switch state {
	case assettag.LifecycleStateActive, assettag.LifecycleStateArchived:
	default:
		return assettag.Tag{}, false
	}
	return assettag.Tag{
		ID:             id,
		TenantID:       assettag.TenantID(m.TenantID),
		InventoryID:    assettag.InventoryID(m.InventoryID),
		Key:            key,
		DisplayName:    displayName,
		Color:          color,
		LifecycleState: state,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}, true
}
