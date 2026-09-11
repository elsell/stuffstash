package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func NotificationToResponse(value ports.NotificationRecord, item asset.Asset, ancestors []ports.NotificationAncestor, incomplete bool) dto.NotificationResponse {
	trail := make([]dto.NotificationAncestor, 0, len(ancestors))
	for _, ancestor := range ancestors {
		trail = append(trail, dto.NotificationAncestor{AssetID: ancestor.AssetID.String(), Title: ancestor.Title, Kind: string(ancestor.Kind)})
	}
	return dto.NotificationResponse{ParentTrail: trail, ParentTrailIncomplete: incomplete, ID: value.ID, AssetID: item.ID.String(), Title: item.Title.String(), ParentAssetID: item.ParentAssetID.String(), CustomAssetTypeID: item.CustomAssetTypeID.String(), ExpirationDate: value.Milestone.Date.Value(), ExpirationPrecision: string(value.Milestone.Date.Precision()), Milestone: string(value.Milestone.Kind), CreatedAt: value.CreatedAt, ReadAt: value.ReadAt}
}
