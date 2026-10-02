package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

// ArchiveSnapshotSources are valid only inside WithArchiveSnapshot's callback.
// They share one database snapshot; none may escape into a background goroutine.
type ArchiveSnapshotSources struct {
	Inventories InventoryRepository
	Assets      AssetRepository
	Tags        AssetTagRepository
	Checkouts   AssetCheckoutRepository
	Attachments AttachmentRepository
	Types       CustomAssetTypeRepository
	Fields      CustomFieldDefinitionRepository
}

type ArchiveSnapshotRepository interface {
	WithArchiveSnapshot(context.Context, tenant.ID, inventory.InventoryID, func(ArchiveSnapshotSources) error) error
}
