package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AssetTagRepository interface {
	AssetTagByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, tagID assettag.ID) (assettag.Tag, bool, error)
	AssetTagByKey(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, key assettag.Key) (assettag.Tag, bool, error)
	ListAssetTags(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AssetTagPageRequest) ([]assettag.Tag, error)
	AssetTagsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) ([]assettag.Tag, error)
	AssetTagsByAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID][]assettag.Tag, error)
}

type AssetTagUnitOfWork interface {
	CreateAssetTag(ctx context.Context, tag assettag.Tag, auditRecord audit.Record) error
	UpdateAssetTag(ctx context.Context, tag assettag.Tag, auditRecord audit.Record) error
	UpdateAssetTagLifecycle(ctx context.Context, tag assettag.Tag, auditRecord audit.Record) error
	SetAssetTags(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, tagIDs []assettag.ID, auditRecord audit.Record) error
}

// AssetEditUnitOfWork persists a direct asset edit, its optional complete tag
// replacement, audit history, and undo snapshot as one transaction.
type AssetEditUnitOfWork interface {
	UpdateAssetAndTags(ctx context.Context, item asset.Asset, tagIDs []assettag.ID, auditRecords []audit.Record, undoableOperation *UndoableOperation) error
}

type AssetTagPageRequest struct {
	AfterTagID assettag.ID
	Limit      int
}
