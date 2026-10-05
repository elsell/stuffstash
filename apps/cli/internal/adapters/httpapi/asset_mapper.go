package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// assetListResponse bypasses nullable.UnmarshalJSON, which rounds untyped
// custom-field numbers. The embedded generated envelope retains metadata.
type assetListResponse struct {
	generated.SuccessEnvelopeListAssetResponse
	Data []generated.AssetResponse `json:"data"`
}

func asset(a generated.AssetResponse) ports.Asset {
	result := ports.Asset{ID: a.Id, TenantID: a.TenantId, InventoryID: a.InventoryId, Title: a.Title, Kind: a.Kind, Description: a.Description, Parent: a.ParentAssetId, Lifecycle: a.LifecycleState, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt, CustomAssetTypeID: a.CustomAssetTypeId, CustomFields: a.CustomFields, PrintJobID: a.PrintJobId, UndoableOperationID: a.UndoableOperationId}
	if value, err := a.Expiration.Get(); err == nil {
		result.Expiration = &ports.Expiration{Date: value.Date, Precision: string(value.Precision)}
	}
	if a.ExpirationContext != nil {
		v := a.ExpirationContext
		result.ExpirationContext = &ports.ExpirationContext{State: string(v.State), TrackingEnabled: v.TrackingEnabled, AdvanceDays: v.AdvanceDays, Timezone: v.Timezone}
	}
	if tags := a.Tags.GetOrEmpty(); tags != nil {
		result.Tags = make([]ports.CompactTag, 0, len(tags))
		for _, v := range tags {
			result.Tags = append(result.Tags, ports.CompactTag{ID: v.Id, Key: v.Key, DisplayName: v.DisplayName, Color: v.Color})
		}
	}
	if a.CurrentCheckout != nil {
		v := currentCheckout(*a.CurrentCheckout)
		result.CurrentCheckout = &v
	}
	if a.PrimaryPhoto != nil {
		v := a.PrimaryPhoto
		result.PrimaryPhoto = &ports.PrimaryPhoto{ID: v.Id, FileName: v.FileName, ContentType: v.ContentType, SizeBytes: v.SizeBytes, Thumbnails: ports.PhotoThumbnails{Small: v.Thumbnails.Small, Medium: v.Thumbnails.Medium, Large: v.Thumbnails.Large}}
	}
	return result
}

func assetQuery(q ports.AssetQuery) *generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsParams {
	p := &generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsParams{Limit: &q.Page.Limit, Cursor: &q.Page.Cursor}
	if q.Lifecycle != "" {
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsParamsLifecycleState(q.Lifecycle)
		p.LifecycleState = &v
	}
	if q.Sort != "" {
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsParamsSort(q.Sort)
		p.Sort = &v
	}
	return p
}
