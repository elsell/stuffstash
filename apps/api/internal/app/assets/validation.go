package assets

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) validatedAssetCustomAssetTypeID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, rawCustomAssetTypeID string) (asset.CustomAssetTypeID, error) {
	return ValidateAssetCustomAssetTypeID(ctx, s.customAssetTypes, tenantID, inventoryID, rawCustomAssetTypeID)
}

func ValidateAssetCustomAssetTypeID(ctx context.Context, repo ports.CustomAssetTypeRepository, tenantID tenant.ID, inventoryID inventory.InventoryID, rawCustomAssetTypeID string) (asset.CustomAssetTypeID, error) {
	if strings.TrimSpace(rawCustomAssetTypeID) == "" {
		return "", nil
	}
	customAssetTypeID, ok := customfield.NewAssetTypeID(rawCustomAssetTypeID)
	if !ok {
		return "", apperrors.ErrInvalidInput
	}
	if repo == nil {
		return "", apperrors.ErrInvalidInput
	}
	types, err := repo.CustomAssetTypesByID(ctx, tenantID, inventoryID, []customfield.AssetTypeID{customAssetTypeID})
	if err != nil {
		return "", err
	}
	if len(types) != 1 {
		return "", apperrors.ErrNotFound
	}
	return asset.CustomAssetTypeID(customAssetTypeID.String()), nil
}

func (s Service) validatedCustomFields(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, customAssetTypeID asset.CustomAssetTypeID, values map[string]any) (asset.CustomFields, error) {
	return ValidateCustomFields(ctx, s.customFields, tenantID, inventoryID, customAssetTypeID, values)
}

func ValidateCustomFields(ctx context.Context, repo ports.CustomFieldDefinitionRepository, tenantID tenant.ID, inventoryID inventory.InventoryID, customAssetTypeID asset.CustomAssetTypeID, values map[string]any) (asset.CustomFields, error) {
	customFields, ok := asset.NewCustomFields(normalizeCustomFieldValues(values))
	if !ok {
		return asset.CustomFields{}, apperrors.ErrInvalidInput
	}
	if customFields.IsEmpty() {
		return customFields, nil
	}
	if repo == nil {
		return asset.CustomFields{}, apperrors.ErrInvalidInput
	}
	definitions, err := repo.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return asset.CustomFields{}, err
	}
	if !customfield.DefinitionSet(definitions).ValidateValuesForAssetType(customFields.Values(), customfield.AssetTypeID(customAssetTypeID.String())) {
		return asset.CustomFields{}, apperrors.ErrInvalidInput
	}
	return customFields, nil
}

func (s Service) validatedParentAssetID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, rawParentAssetID string) (asset.ID, error) {
	return s.validatedParentAssetIDWithPendingParents(ctx, tenantID, inventoryID, assetID, rawParentAssetID, nil)
}

func (s Service) validatedParentAssetIDWithPendingParents(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, rawParentAssetID string, pendingParents map[asset.ID]asset.Kind) (asset.ID, error) {
	if strings.TrimSpace(rawParentAssetID) == "" {
		return asset.ID(""), apperrors.ErrInvalidInput
	}
	parentAssetID, ok := asset.NewID(rawParentAssetID)
	if !ok || parentAssetID == assetID {
		return asset.ID(""), apperrors.ErrInvalidInput
	}
	if pendingKind, ok := pendingParents[parentAssetID]; ok {
		if !pendingKind.CanContainChildren() {
			return asset.ID(""), apperrors.ErrInvalidInput
		}
		return parentAssetID, nil
	}
	parent, found, err := s.assets.AssetByID(ctx, tenantID, inventoryID, parentAssetID)
	if err != nil {
		return asset.ID(""), err
	}
	if !found {
		return asset.ID(""), apperrors.ErrNotFound
	}
	if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
		return asset.ID(""), apperrors.ErrInvalidInput
	}
	for current := parent; current.ParentAssetID.String() != ""; {
		if current.ParentAssetID == assetID {
			return asset.ID(""), apperrors.ErrInvalidInput
		}
		next, found, err := s.assets.AssetByID(ctx, tenantID, inventoryID, current.ParentAssetID)
		if err != nil {
			return asset.ID(""), err
		}
		if !found {
			return asset.ID(""), apperrors.ErrInvalidInput
		}
		current = next
	}
	return parentAssetID, nil
}

func normalizeCustomFieldValues(values map[string]any) map[string]any {
	normalized := map[string]any{}
	for key, value := range values {
		normalized[key] = customfield.NormalizeJSONNumber(value)
	}
	return normalized
}
