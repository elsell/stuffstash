package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
)

func AssetToResponse(item asset.Asset) dto.AssetResponse {
	return dto.AssetResponse{
		ID:                item.ID.String(),
		TenantID:          item.TenantID.String(),
		InventoryID:       item.InventoryID.String(),
		ParentAssetID:     item.ParentAssetID.String(),
		CustomAssetTypeID: item.CustomAssetTypeID.String(),
		Kind:              item.Kind.String(),
		Title:             item.Title.String(),
		Description:       item.Description.String(),
		CustomFields:      item.CustomFields.Values(),
		LifecycleState:    item.LifecycleState.String(),
	}
}

func AssetsToResponse(items []asset.Asset) []dto.AssetResponse {
	data := make([]dto.AssetResponse, 0, len(items))
	for _, item := range items {
		data = append(data, AssetToResponse(item))
	}
	return data
}
