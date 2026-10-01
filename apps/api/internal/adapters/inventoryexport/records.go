package inventoryexport

import (
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

// Explicit projections keep new infrastructure fields out of portable files.
func assetRecord(a ports.InventoryExportAsset) map[string]any {
	attachments := make([]map[string]any, 0, len(a.Attachments))
	for _, m := range a.Attachments {
		attachments = append(attachments, map[string]any{
			"id": m.ID, "fileName": m.FileName, "contentType": m.ContentType, "sizeBytes": m.SizeBytes, "sha256": m.SHA256, "createdAt": m.CreatedAt, "lifecycleState": m.LifecycleState,
		})
	}
	var checkout any
	if c := a.CurrentCheckout; c != nil {
		checkout = map[string]any{"id": c.ID, "state": c.State, "checkedOutAt": c.CheckedOutAt, "checkedOutByPrincipal": c.CheckedOutByPrincipal, "details": c.CheckoutDetails}
	}
	return map[string]any{"id": a.ID, "title": a.Title, "description": a.Description, "kind": a.Kind, "parentAssetId": a.ParentAssetID, "customAssetTypeId": a.CustomAssetTypeID, "lifecycleState": a.LifecycleState, "createdAt": a.CreatedAt.Format(time.RFC3339Nano), "updatedAt": a.UpdatedAt.Format(time.RFC3339Nano), "expirationDate": a.ExpirationDate, "expirationPrecision": a.ExpirationPrecision, "tagIds": a.TagIDs, "customFields": a.CustomFields, "currentCheckout": checkout, "attachments": attachments}
}
func tagRecord(t assettag.Tag) map[string]any {
	return map[string]any{"id": t.ID, "key": t.Key, "displayName": t.DisplayName, "color": t.Color, "lifecycleState": t.LifecycleState, "createdAt": t.CreatedAt, "updatedAt": t.UpdatedAt}
}
func assetTypeRecord(t customfield.AssetType) map[string]any {
	return map[string]any{"id": t.ID, "scope": t.Scope, "key": t.Key, "displayName": t.DisplayName, "description": t.Description, "lifecycleState": t.LifecycleState, "expirationEnabled": t.ExpirationEnabled}
}
func fieldRecord(f customfield.Definition) map[string]any {
	return map[string]any{"id": f.ID, "scope": f.Scope, "key": f.Key, "displayName": f.DisplayName, "type": f.Type, "enumOptions": f.EnumOptions, "applicability": f.Applicability, "customAssetTypeIds": f.CustomAssetTypeIDs, "lifecycleState": f.LifecycleState}
}
