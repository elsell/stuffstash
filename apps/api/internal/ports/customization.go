package ports

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type CustomFieldDefinitionRepository interface {
	CustomFieldDefinitionHasActiveAssetValues(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definition customfield.Definition) (bool, error)
	CustomFieldDefinitionByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error)
	ListTenantCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListInventoryCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error)
}

type CustomFieldDefinitionUnitOfWork interface {
	SaveCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	UpdateCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	UpdateCustomFieldDefinitionLifecycle(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	DeleteCustomFieldDefinition(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID, auditRecord audit.Record) error
}

type CustomFieldDefinitionPageRequest struct {
	AfterDefinitionKey string
	Limit              int
	Lifecycle          CustomizationLifecycleFilter
}

type CustomizationLifecycleFilter string

const (
	CustomizationLifecycleActive   CustomizationLifecycleFilter = "active"
	CustomizationLifecycleArchived CustomizationLifecycleFilter = "archived"
	CustomizationLifecycleAll      CustomizationLifecycleFilter = "all"
)

func ParseCustomizationLifecycleFilter(value string) (CustomizationLifecycleFilter, bool) {
	switch CustomizationLifecycleFilter(strings.TrimSpace(value)) {
	case "", CustomizationLifecycleActive:
		return CustomizationLifecycleActive, true
	case CustomizationLifecycleArchived:
		return CustomizationLifecycleArchived, true
	case CustomizationLifecycleAll:
		return CustomizationLifecycleAll, true
	default:
		return "", false
	}
}

func (f CustomizationLifecycleFilter) String() string {
	return string(f)
}

func (f CustomizationLifecycleFilter) Includes(value string) bool {
	if value == "" {
		value = CustomizationLifecycleActive.String()
	}
	if f == "" {
		return value == CustomizationLifecycleActive.String()
	}
	return f == CustomizationLifecycleAll || f.String() == value
}

type CustomAssetTypeRepository interface {
	CustomAssetTypeHasActiveReferences(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (bool, error)
	CustomAssetTypeByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error)
	ListTenantCustomAssetTypes(ctx context.Context, tenantID tenant.ID, page CustomAssetTypePageRequest) ([]customfield.AssetType, error)
	ListInventoryCustomAssetTypes(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomAssetTypePageRequest) ([]customfield.AssetType, error)
	CustomAssetTypesByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error)
}

type CustomAssetTypeUnitOfWork interface {
	SaveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	UpdateCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	ArchiveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	RestoreCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	DeleteCustomAssetType(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID, auditRecord audit.Record) error
}

type CustomAssetTypePageRequest struct {
	AfterAssetTypeKey string
	Limit             int
	Lifecycle         CustomizationLifecycleFilter
}
