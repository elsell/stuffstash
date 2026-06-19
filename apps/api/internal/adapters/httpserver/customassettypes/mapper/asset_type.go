package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

func AssetTypeToResponse(assetType customfield.AssetType) dto.AssetTypeResponse {
	return dto.AssetTypeResponse{
		ID:          assetType.ID.String(),
		TenantID:    assetType.TenantID.String(),
		InventoryID: assetType.InventoryID.String(),
		Scope:       assetType.Scope.String(),
		Key:         assetType.Key.String(),
		DisplayName: assetType.DisplayName.String(),
		Description: assetType.Description.String(),
	}
}

func AssetTypesToResponse(assetTypes []customfield.AssetType) []dto.AssetTypeResponse {
	data := make([]dto.AssetTypeResponse, 0, len(assetTypes))
	for _, assetType := range assetTypes {
		data = append(data, AssetTypeToResponse(assetType))
	}
	return data
}
