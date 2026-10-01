package mcpserver

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"time"
)

func assetValue(item asset.Asset) map[string]any {
	return map[string]any{"id": item.ID.String(), "tenantId": item.TenantID.String(), "inventoryId": item.InventoryID.String(), "parentAssetId": item.ParentAssetID.String(), "kind": item.Kind.String(), "title": item.Title.String(), "description": item.Description.String(), "lifecycleState": item.LifecycleState.String(), "customAssetTypeId": item.CustomAssetTypeID.String(), "customFields": item.CustomFields.Values(), "expirationDate": item.Expiration.Value(), "expirationPrecision": string(item.Expiration.Precision())}
}
func tagValues(tags []assettag.Tag) []any {
	values := []any{}
	for _, tag := range tags {
		values = append(values, map[string]any{"id": tag.ID.String(), "name": tag.DisplayName.String(), "color": string(tag.Color)})
	}
	return values
}
func checkoutValue(item asset.Checkout) map[string]any {
	value := map[string]any{"id": string(item.ID), "assetId": item.AssetID.String(), "state": string(item.State), "checkedOutAt": item.CheckedOutAt.UTC().Format(time.RFC3339Nano), "checkoutDetails": item.CheckoutDetails.String()}
	if !item.ReturnedAt.IsZero() {
		value["returnedAt"] = item.ReturnedAt.UTC().Format(time.RFC3339Nano)
		value["returnDetails"] = item.ReturnDetails.String()
	}
	return value
}
