package app

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/audithistory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AssetActivityView = audithistory.AssetActivityView
type ListAssetActivityInput = audithistory.ListAssetActivityInput
type ListAssetActivityResult = audithistory.ListAssetActivityResult

const (
	AssetActivityViewChanges = audithistory.AssetActivityViewChanges
	AssetActivityViewAll     = audithistory.AssetActivityViewAll
)

func (a App) ListAssetActivity(ctx context.Context, input ListAssetActivityInput) (ListAssetActivityResult, error) {
	return a.auditHistoryService().ListAssetActivity(ctx, input)
}
func (a App) projectAssetActivityEntry(ctx context.Context, input ListAssetActivityInput, record audit.Record, canUndo bool) audit.AssetActivityEntry {
	return a.auditHistoryService().ProjectAssetActivityEntry(ctx, input, record, canUndo)
}
func encodeAssetActivityCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, view AssetActivityView, record audit.Record) *string {
	return audithistory.EncodeAssetActivityCursor(tenantID, inventoryID, assetID, view, record)
}
func decodeAssetActivityCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, view AssetActivityView, cursor string) (time.Time, audit.ID, error) {
	return audithistory.DecodeAssetActivityCursor(tenantID, inventoryID, assetID, view, cursor)
}
