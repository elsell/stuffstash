package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"gorm.io/gorm"
	"time"
)

type customFieldDefinitionModel struct {
	ID            string          `gorm:"primaryKey;size:26"`
	TenantID      string          `gorm:"not null;size:26;index"`
	Tenant        tenantModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID   *string         `gorm:"size:26;index"`
	Inventory     *inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	Scope         string          `gorm:"not null;size:32;index;check:chk_custom_field_definitions_scope,scope IN ('tenant','inventory')"`
	CursorKey     string          `gorm:"not null;size:32;index"`
	FieldKey      string          `gorm:"not null;size:80;index"`
	DisplayName   string          `gorm:"not null;size:120"`
	FieldType     string          `gorm:"not null;size:32;check:chk_custom_field_definitions_field_type,field_type IN ('text','number','boolean','date','url','enum')"`
	EnumOptions   string          `gorm:"type:jsonb;not null;default:'[]'"`
	Applicability string          `gorm:"not null;size:32;default:'all_assets';check:chk_custom_field_definitions_applicability,applicability IN ('all_assets','custom_asset_types')"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type customFieldDefinitionAssetTypeModel struct {
	CustomFieldDefinitionID string                     `gorm:"primaryKey;size:26"`
	CustomFieldDefinition   customFieldDefinitionModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CustomFieldDefinitionID;references:ID"`
	CustomAssetTypeID       string                     `gorm:"primaryKey;size:26"`
	CustomAssetType         customAssetTypeModel       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:CustomAssetTypeID;references:ID"`
	TenantID                string                     `gorm:"not null;size:26;index"`
	Tenant                  tenantModel                `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID             *string                    `gorm:"size:26;index"`
	Inventory               *inventoryModel            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	CreatedAt               time.Time
}

func (customFieldDefinitionModel) TableName() string {
	return "custom_field_definitions"
}

func (customFieldDefinitionAssetTypeModel) TableName() string {
	return "custom_field_definition_asset_types"
}

func (m *customFieldDefinitionModel) BeforeSave(*gorm.DB) error {
	scope := customfield.Scope(m.Scope)
	prefix := "1:"
	if scope == customfield.ScopeTenant {
		prefix = "0:"
	}
	m.CursorKey = prefix + m.ID
	return nil
}

func (m customFieldDefinitionModel) toDomain(customAssetTypeIDs []customfield.AssetTypeID) (customfield.Definition, bool) {
	id, ok := customfield.NewID(m.ID)
	if !ok {
		return customfield.Definition{}, false
	}
	key, ok := customfield.NewKey(m.FieldKey)
	if !ok {
		return customfield.Definition{}, false
	}
	displayName, ok := customfield.NewDisplayName(m.DisplayName)
	if !ok {
		return customfield.Definition{}, false
	}
	fieldType, ok := customfield.NewFieldType(m.FieldType)
	if !ok {
		return customfield.Definition{}, false
	}
	scope := customfield.Scope(m.Scope)
	inventoryID := customfield.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = customfield.InventoryID(*m.InventoryID)
	}
	var rawOptions []string
	if err := json.Unmarshal([]byte(m.EnumOptions), &rawOptions); err != nil {
		return customfield.Definition{}, false
	}
	applicability, ok := customfield.NewApplicability(m.Applicability)
	if !ok {
		return customfield.Definition{}, false
	}
	options := make([]customfield.Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := customfield.NewKey(raw)
		if !ok {
			return customfield.Definition{}, false
		}
		options = append(options, option)
	}
	return customfield.NewDefinition(
		id,
		customfield.TenantID(m.TenantID),
		inventoryID,
		scope,
		key,
		displayName,
		fieldType,
		options,
		applicability,
		customAssetTypeIDs,
	)
}
