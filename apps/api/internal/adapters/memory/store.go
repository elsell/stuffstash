package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
)

type Store struct {
	mu               sync.RWMutex
	tenants          map[tenant.ID]tenant.Tenant
	inventories      map[inventory.InventoryID]inventory.Inventory
	accessGrants     map[string]ports.InventoryAccessGrant
	invitations      map[string]ports.InventoryAccessInvitation
	customAssetTypes map[customfield.AssetTypeID]customfield.AssetType
	customFields     map[customfield.ID]customfield.Definition
	assets           map[asset.ID]asset.Asset
	attachments      map[media.ID]media.Attachment
	blobs            map[media.StorageKey][]byte
	auditRecords     map[audit.ID]audit.Record
	outbox           map[string]ports.AuthorizationOutboxEvent
}

func NewStore() *Store {
	return &Store{
		tenants:          map[tenant.ID]tenant.Tenant{},
		inventories:      map[inventory.InventoryID]inventory.Inventory{},
		accessGrants:     map[string]ports.InventoryAccessGrant{},
		invitations:      map[string]ports.InventoryAccessInvitation{},
		customAssetTypes: map[customfield.AssetTypeID]customfield.AssetType{},
		customFields:     map[customfield.ID]customfield.Definition{},
		assets:           map[asset.ID]asset.Asset{},
		attachments:      map[media.ID]media.Attachment{},
		blobs:            map[media.StorageKey][]byte{},
		auditRecords:     map[audit.ID]audit.Record{},
		outbox:           map[string]ports.AuthorizationOutboxEvent{},
	}
}
