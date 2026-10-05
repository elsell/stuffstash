package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func searchAsset(v generated.AssetSummary) ports.SearchAsset {
	r := ports.SearchAsset{ID: v.Id, InventoryID: v.InventoryId, Title: v.Title, Kind: v.Kind, Description: v.Description, Parent: v.ParentAssetId, Lifecycle: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, CustomAssetTypeID: v.CustomAssetTypeId, CustomFields: v.CustomFields}
	if v.Expiration != nil {
		r.Expiration = &ports.Expiration{Date: v.Expiration.Date, Precision: string(v.Expiration.Precision)}
	}
	if c := v.ExpirationContext; c != nil {
		r.ExpirationContext = &ports.ExpirationContext{State: string(c.State), TrackingEnabled: c.TrackingEnabled, AdvanceDays: c.AdvanceDays, Timezone: c.Timezone}
	}
	if tags := v.Tags.GetOrEmpty(); tags != nil {
		r.Tags = make([]ports.CompactTag, 0, len(tags))
		for _, tag := range tags {
			r.Tags = append(r.Tags, ports.CompactTag{ID: tag.Id, Key: tag.Key, DisplayName: tag.DisplayName, Color: tag.Color})
		}
	}
	if c := v.CurrentCheckout; c != nil {
		r.CurrentCheckout = &ports.CurrentCheckout{ID: c.Id, State: c.State, CheckedOutAt: c.CheckedOutAt, CheckedOutByPrincipalID: c.CheckedOutByPrincipalId}
		if c.CheckedOutByPrincipal != nil {
			r.CurrentCheckout.CheckedOutByPrincipal = &ports.Principal{ID: c.CheckedOutByPrincipal.Id, Email: c.CheckedOutByPrincipal.Email}
		}
	}
	if p := v.PrimaryPhoto; p != nil {
		r.PrimaryPhoto = &ports.PrimaryPhoto{ID: p.Id, FileName: p.FileName, ContentType: p.ContentType, SizeBytes: p.SizeBytes, Thumbnails: ports.PhotoThumbnails{Small: p.Thumbnails.Small, Medium: p.Thumbnails.Medium, Large: p.Thumbnails.Large}}
	}
	return r
}
