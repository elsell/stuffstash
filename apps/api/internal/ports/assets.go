package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AssetRepository interface {
	AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error)
	AssetHasActiveChildren(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (bool, error)
	ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AssetListPageRequest) ([]asset.Asset, error)
}

type AssetUnitOfWork interface {
	CreateAsset(ctx context.Context, asset asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	UpdateAsset(ctx context.Context, asset asset.Asset, auditRecords []audit.Record, undoableOperation *UndoableOperation) error
	UpdateAssetLifecycle(ctx context.Context, asset asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	DeleteAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, auditRecord audit.Record) error
}

type UndoableOperationRepository interface {
	UndoableOperationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, operationID string) (UndoableOperation, bool, error)
	ApplyAssetUndoableOperation(ctx context.Context, operationID string, direction UndoableOperationDirection, expectedCurrent asset.Asset, resulting asset.Asset, auditRecord audit.Record) (UndoableOperation, asset.Asset, error)
}

type UndoableOperationStatus string

const (
	UndoableOperationAvailable UndoableOperationStatus = "available"
	UndoableOperationUndone    UndoableOperationStatus = "undone"
	UndoableOperationRedone    UndoableOperationStatus = "redone"
)

type UndoableOperationDirection string

const (
	UndoableOperationDirectionUndo UndoableOperationDirection = "undo"
	UndoableOperationDirectionRedo UndoableOperationDirection = "redo"
)

type UndoableOperation struct {
	ID                string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	PrincipalID       identity.PrincipalID
	Source            audit.Source
	TargetType        audit.TargetType
	TargetID          string
	OriginalAction    audit.Action
	Status            UndoableOperationStatus
	CreatedAt         time.Time
	LastAppliedAt     time.Time
	BeforeAsset       *asset.Asset
	AfterAsset        asset.Asset
	UndoAuditRecordID audit.ID
	RedoAuditRecordID audit.ID
}

type AssetLifecycleFilter string

const (
	AssetLifecycleFilterActive   AssetLifecycleFilter = "active"
	AssetLifecycleFilterArchived AssetLifecycleFilter = "archived"
	AssetLifecycleFilterAll      AssetLifecycleFilter = "all"
)

type AssetListPageRequest struct {
	AfterAssetID    asset.ID
	AfterUpdatedAt  time.Time
	Limit           int
	LifecycleFilter AssetLifecycleFilter
	Sort            AssetListSort
}

type AssetListSort string

const (
	AssetListSortIDAsc       AssetListSort = "id_asc"
	AssetListSortUpdatedDesc AssetListSort = "updated_desc"
)
