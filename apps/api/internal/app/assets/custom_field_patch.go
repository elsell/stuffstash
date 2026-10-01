package assets

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// ApplyCustomFieldPatch validates even explicit clears before merging. Callers
// supply the same current snapshot they will use for optimistic persistence.
func ApplyCustomFieldPatch(ctx context.Context, repo ports.CustomFieldDefinitionRepository, tenantID tenant.ID, inventoryID inventory.InventoryID, typeID asset.CustomAssetTypeID, current, patch map[string]any) (map[string]any, error) {
	if repo == nil {
		return nil, apperrors.ErrInvalidInput
	}
	definitions, err := repo.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return nil, err
	}
	byKey := map[string]customfield.Definition{}
	for _, definition := range definitions {
		if definition.IsActive() && definition.AppliesTo(customfield.AssetTypeID(typeID)) {
			byKey[definition.Key.String()] = definition
		}
	}
	merged := make(map[string]any, len(current)+len(patch))
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range patch {
		definition, ok := byKey[key]
		if !ok || value != nil && !definition.ValidValue(value) {
			return nil, apperrors.ErrInvalidInput
		}
		if value == nil {
			delete(merged, key)
		} else {
			merged[key] = value
		}
	}
	return merged, nil
}
