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
	CreateAssetWithParentPromotion(ctx context.Context, promotedParent asset.Asset, parentAuditRecord audit.Record, asset asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	UpdateAsset(ctx context.Context, asset asset.Asset, auditRecords []audit.Record, undoableOperation *UndoableOperation) error
	UpdateAssetLifecycle(ctx context.Context, asset asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	CheckOutAsset(ctx context.Context, checkout asset.Checkout, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	ReturnAsset(ctx context.Context, expectedCurrent asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	UpdateAssetCheckoutReturnDetails(ctx context.Context, expectedCurrent asset.Checkout, updated asset.Checkout, auditRecord audit.Record) error
	DeleteAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, auditRecord audit.Record) error
}

type AssetCheckoutRepository interface {
	CurrentAssetCheckout(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Checkout, bool, error)
	CurrentAssetCheckouts(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID]asset.Checkout, error)
	AssetCheckoutByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, checkoutID asset.CheckoutID) (asset.Checkout, bool, error)
	ListAssetCheckoutHistory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page AssetCheckoutHistoryPageRequest) ([]asset.Checkout, error)
	ListCheckedOutAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CheckedOutAssetsPageRequest) ([]CheckedOutAsset, error)
	HasLaterCheckout(ctx context.Context, checkout asset.Checkout) (bool, error)
}

type UndoableOperationRepository interface {
	UndoableOperationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, operationID string) (UndoableOperation, bool, error)
	ApplyAssetUndoableOperation(ctx context.Context, operationID string, direction UndoableOperationDirection, expectedCurrent asset.Asset, resulting asset.Asset, auditRecord audit.Record) (UndoableOperation, asset.Asset, error)
	ApplyAssetCheckoutUndoableOperation(ctx context.Context, operationID string, direction UndoableOperationDirection, expectedCurrent asset.Checkout, resulting asset.Checkout, auditRecord audit.Record) (UndoableOperation, asset.Checkout, error)
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
	BeforeCheckout    *asset.Checkout
	AfterCheckout     *asset.Checkout
	UndoAuditRecordID audit.ID
	RedoAuditRecordID audit.ID
}

type AssetCheckoutHistoryPageRequest struct {
	AfterCheckoutID   asset.CheckoutID
	AfterCheckedOutAt time.Time
	Limit             int
}

type CheckedOutAssetsPageRequest struct {
	AfterAssetID      asset.ID
	AfterCheckedOutAt time.Time
	Limit             int
}

type CheckedOutAsset struct {
	Asset    asset.Asset
	Checkout asset.Checkout
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
