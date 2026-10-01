package assets

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func (s Service) validateUndoableAssetResult(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, item asset.Asset, assigningType bool) error {
	customAssetTypeID := item.CustomAssetTypeID
	if customAssetTypeID.String() != "" {
		err := s.ensureSnapshotCustomAssetTypeExists(ctx, tenantID, inventoryID, customAssetTypeID, assigningType)
		if err != nil {
			return err
		}
	}
	if _, err := ValidateCustomFields(ctx, s.customFields, tenantID, inventoryID, customAssetTypeID, item.CustomFields.Values()); err != nil {
		return err
	}
	return nil
}

func (s Service) ensureSnapshotCustomAssetTypeExists(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, customAssetTypeID asset.CustomAssetTypeID, requireActive bool) error {
	parsed, ok := customfield.NewAssetTypeID(customAssetTypeID.String())
	if !ok {
		return apperrors.ErrInvalidInput
	}
	if s.customAssetTypes == nil {
		return apperrors.ErrInvalidInput
	}
	kind, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, tenantID, inventoryID, parsed)
	if err != nil {
		return err
	}
	if !found {
		return apperrors.ErrNotFound
	}
	if requireActive && !kind.IsActive() {
		return apperrors.ErrInvalidInput
	}

	return nil
}
