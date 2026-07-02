package assets

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAssetInput struct {
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
}

type ListAssetsInput struct {
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

type AssetParentUpdate struct {
	Present bool
	Null    bool
	Value   string
}

type UpdateAssetInput struct {
	Principal     identity.Principal
	Source        audit.Source
	RequestID     string
	TenantID      tenant.ID
	InventoryID   inventory.InventoryID
	AssetID       asset.ID
	Title         *string
	Description   *string
	ParentAssetID AssetParentUpdate
	CustomFields  map[string]any
}

type UpdateAssetLifecycleInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}

type ListAssetsResult struct {
	Items         []asset.Asset
	PrimaryPhotos map[ports.AttachmentAssetReference]media.Attachment
	Limit         int
	NextCursor    *string
	HasMore       bool
}

type GetAssetResult struct {
	Item         asset.Asset
	PrimaryPhoto *media.Attachment
}
