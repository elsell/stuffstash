package assets

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func LifecycleFilter(value string) (ports.AssetLifecycleFilter, error) {
	switch strings.TrimSpace(value) {
	case "":
		return ports.AssetLifecycleFilterActive, nil
	case string(ports.AssetLifecycleFilterActive):
		return ports.AssetLifecycleFilterActive, nil
	case string(ports.AssetLifecycleFilterArchived):
		return ports.AssetLifecycleFilterArchived, nil
	case string(ports.AssetLifecycleFilterAll):
		return ports.AssetLifecycleFilterAll, nil
	default:
		return "", apperrors.ErrInvalidInput
	}
}

func encodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, id asset.ID) *string {
	return encodePageCursor("assets", tenantID.String()+":"+inventoryID.String()+":"+string(lifecycleFilter), id.String())
}

func decodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, cursor string) (asset.ID, error) {
	decoded, err := decodePageCursor("assets", tenantID.String()+":"+inventoryID.String()+":"+string(lifecycleFilter), cursor)
	if err != nil {
		return asset.ID(""), err
	}
	if decoded == "" {
		return asset.ID(""), nil
	}
	id, ok := asset.NewID(decoded)
	if !ok {
		return asset.ID(""), apperrors.ErrInvalidInput
	}
	return id, nil
}

func encodePageCursor(collection string, scope string, lastID string) *string {
	return appsupport.EncodePageCursor(collection, scope, lastID)
}

func decodePageCursor(collection string, scope string, cursor string) (string, error) {
	return appsupport.DecodePageCursor(collection, scope, cursor)
}
