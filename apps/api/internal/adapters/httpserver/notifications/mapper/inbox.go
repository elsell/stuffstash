package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func NotificationToResponse(value ports.NotificationRecord, item asset.Asset) dto.NotificationResponse {
	return dto.NotificationResponse{ID: value.ID, AssetID: item.ID.String(), Title: item.Title.String(), ParentAssetID: item.ParentAssetID.String(), CustomAssetTypeID: item.CustomAssetTypeID.String(), ExpirationDate: value.Milestone.Date.Value(), ExpirationPrecision: string(value.Milestone.Date.Precision()), Milestone: string(value.Milestone.Kind), CreatedAt: value.CreatedAt, ReadAt: value.ReadAt}
}
