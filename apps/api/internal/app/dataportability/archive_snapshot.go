package dataportability

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// CaptureArchiveMetadata reuses the bounded export collector inside a single
// repository snapshot. Authorization belongs to the calling job service.
func CaptureArchiveMetadata(ctx context.Context, snapshots ports.ArchiveSnapshotRepository, tenantID tenant.ID, inventoryID inventory.InventoryID, now time.Time, maxRecords int) (ports.InventoryExportDocument, error) {
	result := ports.InventoryExportDocument{}
	if snapshots == nil || maxRecords <= 0 || tenantID.String() == "" || inventoryID.String() == "" || now.IsZero() {
		return result, apperrors.ErrInvalidInput
	}
	err := snapshots.WithArchiveSnapshot(ctx, tenantID, inventoryID, func(repos ports.ArchiveSnapshotSources) error {
		item, found, err := repos.Inventories.InventoryByID(ctx, tenantID, inventoryID)
		if err != nil {
			return err
		}
		if !found || item.TenantID.String() != tenantID.String() || item.ID != inventoryID {
			return apperrors.ErrNotFound
		}
		result = ports.InventoryExportDocument{SchemaVersion: 2, ExportedAt: now, TenantID: tenantID.String(), InventoryID: inventoryID.String(), InventoryName: item.Name.String()}
		collector := New(Dependencies{Inventories: repos.Inventories, Assets: repos.Assets, Tags: repos.Tags, Checkouts: repos.Checkouts, Attachments: repos.Attachments, Types: repos.Types, Fields: repos.Fields, MaxRecords: maxRecords})
		input := ExportInput{TenantID: tenantID, InventoryID: inventoryID}
		if err = collector.collectSchema(ctx, input, &result); err != nil {
			return err
		}
		return collector.collectAssets(ctx, input, &result)
	})
	if err != nil {
		return ports.InventoryExportDocument{}, err
	}
	if err = ValidateArchiveDocument(ctx, result, maxRecords); err != nil {
		return ports.InventoryExportDocument{}, err
	}
	return result, nil
}
