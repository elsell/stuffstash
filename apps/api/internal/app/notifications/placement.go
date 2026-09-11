package notifications

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
)

const maxNotificationAncestors = 128

type placementLookup struct {
	item  asset.Asset
	found bool
}
type placementCache map[asset.ID]placementLookup

func (s Service) enrichPlacement(ctx context.Context, input ScopeInput, view *NotificationView, cache placementCache) error {
	view.ParentTrail = make([]ports.NotificationAncestor, 0)
	next := view.Asset.ParentAssetID
	seen := map[asset.ID]bool{view.Asset.ID: true}
	for next != "" {
		if err := ctx.Err(); err != nil {
			return err
		}
		if seen[next] || len(view.ParentTrail) == maxNotificationAncestors {
			view.ParentTrailIncomplete = true
			break
		}
		seen[next] = true
		entry, ok := cache[next]
		if !ok {
			item, found, err := s.deps.Assets.AssetByID(ctx, input.TenantID, input.InventoryID, next)
			if err != nil {
				return err
			}
			entry = placementLookup{item: item, found: found}
			cache[next] = entry
		}
		item := entry.item
		if !entry.found || item.ID != next || item.TenantID != asset.TenantID(input.TenantID) || item.InventoryID != asset.InventoryID(input.InventoryID) || item.LifecycleState != asset.LifecycleStateActive {
			view.ParentTrailIncomplete = true
			break
		}
		view.ParentTrail = append(view.ParentTrail, ports.NotificationAncestor{AssetID: item.ID, Title: item.Title.String(), Kind: item.Kind})
		next = item.ParentAssetID
	}
	slices.Reverse(view.ParentTrail)
	return nil
}
