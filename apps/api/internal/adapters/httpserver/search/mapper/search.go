package mapper

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/search/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func AssetSearchResultsToResponse(results []ports.AssetSearchResult) []dto.AssetSearchResultResponse {
	data := make([]dto.AssetSearchResultResponse, 0, len(results))
	for _, result := range results {
		data = append(data, AssetSearchResultToResponse(result))
	}
	return data
}

func AssetSearchResultToResponse(result ports.AssetSearchResult) dto.AssetSearchResultResponse {
	return dto.AssetSearchResultResponse{
		Type:     result.Type.String(),
		TenantID: result.TenantID.String(),
		Inventory: dto.InventorySummary{
			ID:   result.Inventory.ID.String(),
			Name: result.Inventory.Name.String(),
		},
		Asset: dto.AssetSummary{
			ID:                result.Asset.ID.String(),
			InventoryID:       result.Asset.InventoryID.String(),
			ParentAssetID:     result.Asset.ParentAssetID.String(),
			CustomAssetTypeID: result.Asset.CustomAssetTypeID.String(),
			Kind:              result.Asset.Kind.String(),
			Title:             result.Asset.Title.String(),
			Description:       result.Asset.Description.String(),
			CustomFields:      result.Asset.CustomFields.Values(),
			LifecycleState:    result.Asset.LifecycleState.String(),
			CreatedAt:         result.Asset.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt:         result.Asset.UpdatedAt.UTC().Format(time.RFC3339Nano),
		},
		Matches: searchMatchesToResponse(result.Matches),
	}
}

func searchMatchesToResponse(matches []search.Match) []dto.SearchMatch {
	data := make([]dto.SearchMatch, 0, len(matches))
	for _, match := range matches {
		data = append(data, dto.SearchMatch{
			Field: match.Field.String(),
			Value: match.Value,
		})
	}
	return data
}
