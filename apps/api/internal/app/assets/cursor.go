package assets

import (
	"encoding/json"
	"strings"
	"time"

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

func AssetSort(value string) (ports.AssetListSort, error) {
	switch strings.TrimSpace(value) {
	case "":
		return ports.AssetListSortIDAsc, nil
	case string(ports.AssetListSortIDAsc):
		return ports.AssetListSortIDAsc, nil
	case string(ports.AssetListSortUpdatedDesc):
		return ports.AssetListSortUpdatedDesc, nil
	default:
		return "", apperrors.ErrInvalidInput
	}
}

type assetCursorPosition struct {
	AssetID   asset.ID
	UpdatedAt time.Time
}

type assetUpdatedCursorPayload struct {
	UpdatedAt string `json:"updatedAt"`
	AssetID   string `json:"assetId"`
}

func encodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, sort ports.AssetListSort, item asset.Asset) *string {
	if sort == ports.AssetListSortUpdatedDesc {
		payload, err := json.Marshal(assetUpdatedCursorPayload{
			UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
			AssetID:   item.ID.String(),
		})
		if err != nil {
			return nil
		}
		return encodePageCursor("assets", assetCursorScope(tenantID, inventoryID, lifecycleFilter, sort), string(payload))
	}
	return encodePageCursor("assets", assetCursorScope(tenantID, inventoryID, lifecycleFilter, sort), item.ID.String())
}

func decodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, sort ports.AssetListSort, cursor string) (assetCursorPosition, error) {
	decoded, err := decodePageCursor("assets", assetCursorScope(tenantID, inventoryID, lifecycleFilter, sort), cursor)
	if err != nil {
		return assetCursorPosition{}, err
	}
	if decoded == "" {
		return assetCursorPosition{}, nil
	}
	if sort == ports.AssetListSortUpdatedDesc {
		var payload assetUpdatedCursorPayload
		if err := json.Unmarshal([]byte(decoded), &payload); err != nil {
			return assetCursorPosition{}, apperrors.ErrInvalidInput
		}
		id, ok := asset.NewID(payload.AssetID)
		if !ok {
			return assetCursorPosition{}, apperrors.ErrInvalidInput
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
		if err != nil {
			return assetCursorPosition{}, apperrors.ErrInvalidInput
		}
		return assetCursorPosition{AssetID: id, UpdatedAt: updatedAt.UTC()}, nil
	}
	id, ok := asset.NewID(decoded)
	if !ok {
		return assetCursorPosition{}, apperrors.ErrInvalidInput
	}
	return assetCursorPosition{AssetID: id}, nil
}

func assetCursorScope(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, sort ports.AssetListSort) string {
	return tenantID.String() + ":" + inventoryID.String() + ":" + string(lifecycleFilter) + ":" + string(sort)
}

func encodePageCursor(collection string, scope string, lastID string) *string {
	return appsupport.EncodePageCursor(collection, scope, lastID)
}

func decodePageCursor(collection string, scope string, cursor string) (string, error) {
	return appsupport.DecodePageCursor(collection, scope, cursor)
}
