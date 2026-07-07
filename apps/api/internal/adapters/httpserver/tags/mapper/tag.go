package mapper

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tags/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
)

func AssetTagToResponse(tag assettag.Tag) dto.AssetTagResponse {
	return dto.AssetTagResponse{
		ID:             tag.ID.String(),
		TenantID:       tag.TenantID.String(),
		InventoryID:    tag.InventoryID.String(),
		Key:            tag.Key.String(),
		DisplayName:    tag.DisplayName.String(),
		Color:          tag.Color.String(),
		LifecycleState: tag.LifecycleState.String(),
		CreatedAt:      tag.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:      tag.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func AssetTagsToResponse(tags []assettag.Tag) []dto.AssetTagResponse {
	data := make([]dto.AssetTagResponse, 0, len(tags))
	for _, tag := range tags {
		data = append(data, AssetTagToResponse(tag))
	}
	return data
}
