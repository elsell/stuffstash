package ports

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type PreparedCreateAsset struct {
	Asset                 asset.Asset
	AuditRecord           audit.Record
	PromotedParent        *asset.Asset
	ParentPromotionRecord *audit.Record
	UndoableOperation     UndoableOperation
}
type PreparedUpdateAsset struct {
	PreviousAsset     asset.Asset
	Asset             asset.Asset
	AuditRecords      []audit.Record
	UndoableOperation *UndoableOperation
	ReplaceTags       bool
	TagIDs            []assettag.ID
}
type PreparedUpdateAssetLifecycle struct {
	PreviousAsset     asset.Asset
	Asset             asset.Asset
	AuditRecord       audit.Record
	UndoableOperation UndoableOperation
	EventName         EventName
	EventMessage      string
}
type CreateAssetInput struct {
	Expiration        *ExpirationInput
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	Kind              string
	Title             string
	Description       string
	ParentAssetID     string
	CustomAssetTypeID string
	CustomFields      map[string]any
	TagIDs            []string
}
type AssetParentUpdate struct {
	Present bool
	Null    bool
	Value   string
}
type UpdateAssetInput struct {
	CustomAssetTypeID *string
	Expiration        ExpirationUpdate
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	AssetID           asset.ID
	Title             *string
	Description       *string
	ParentAssetID     AssetParentUpdate
	CustomFields      map[string]any
	CustomFieldPatch  map[string]any
	TagIDs            *[]string
}
type UpdateAssetLifecycleInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}
type CheckoutAssetInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	Details     string
}
type ReturnAssetInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	Details     string
}
type PreparedCheckoutOperation struct {
	Checkout          asset.Checkout
	ExpectedCurrent   *asset.Checkout
	AuditRecord       audit.Record
	UndoableOperation UndoableOperation
}
type ExpirationInput struct {
	Date      string
	Precision string
}
type ExpirationUpdate struct {
	Present bool
	Value   *ExpirationInput
}
