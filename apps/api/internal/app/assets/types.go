package assets

import (
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAssetInput = ports.CreateAssetInput

type ListAssetsInput struct {
	Parent         ports.AssetParentFilter
	Principal      identity.Principal
	Source         audit.Source
	RequestID      string
	TenantID       tenant.ID
	InventoryID    inventory.InventoryID
	Limit          int
	Cursor         string
	LifecycleState string
	Sort           string
}

type GetAssetInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}

type AssetParentUpdate = ports.AssetParentUpdate

type UpdateAssetInput = ports.UpdateAssetInput

type AssetMutationResult struct {
	Asset               asset.Asset
	UndoableOperationID string
}

type UpdateAssetLifecycleInput = ports.UpdateAssetLifecycleInput

type CheckoutAssetInput = ports.CheckoutAssetInput

type ReturnAssetInput = ports.ReturnAssetInput

type UpdateReturnedCheckoutDetailsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	CheckoutID  asset.CheckoutID
	Details     string
}

type ListAssetCheckoutHistoryInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	Limit       int
	Cursor      string
}

type ListCheckedOutAssetsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type ListAssetsResult struct {
	ExpirationContexts map[ports.AttachmentAssetReference]expirationapp.Description
	Items              []asset.Asset
	Tags               map[asset.ID][]assettag.Tag
	PrimaryPhotos      map[ports.AttachmentAssetReference]media.Attachment
	Checkouts          map[asset.ID]asset.Checkout
	Limit              int
	NextCursor         *string
	HasMore            bool
}

type AssetCheckoutHistoryResult struct {
	Items      []asset.Checkout
	Limit      int
	NextCursor *string
	HasMore    bool
}

type CheckedOutAssetsResult struct {
	Items         []ports.CheckedOutAsset
	PrimaryPhotos map[ports.AttachmentAssetReference]media.Attachment
	Limit         int
	NextCursor    *string
	HasMore       bool
}

type PreparedCheckoutOperation = ports.PreparedCheckoutOperation

type CheckoutOperationResult struct {
	Checkout            asset.Checkout
	UndoableOperationID string
}

type GetAssetResult struct {
	Item            asset.Asset
	Tags            []assettag.Tag
	PrimaryPhoto    *media.Attachment
	CurrentCheckout *asset.Checkout
}
