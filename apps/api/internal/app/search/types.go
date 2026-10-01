package search

import (
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type SearchAssetsInput struct {
	Principal         identity.Principal
	TenantID          tenant.ID
	InventoryIDs      []inventory.InventoryID
	Source            audit.Source
	RequestID         string
	Query             string
	Mode              string
	TagIDs            []assettag.ID
	CustomAssetTypeID string
	LifecycleState    string
	CheckoutState     string
	Limit             int
	Cursor            string
}

type SearchAssetsResult struct {
	AuthorizedInventoryIDs []inventory.InventoryID
	Items                  []ports.AssetSearchResult
	PrimaryPhotos          map[ports.AttachmentAssetReference]media.Attachment
	Limit                  int
	NextCursor             *string
	HasMore                bool
}
