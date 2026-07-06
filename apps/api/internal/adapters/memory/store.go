package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
)

type Store struct {
	mu               sync.RWMutex
	users            map[identity.PrincipalID]identity.User
	tenants          map[tenant.ID]tenant.Tenant
	inventories      map[inventory.InventoryID]inventory.Inventory
	accessGrants     map[string]ports.InventoryAccessGrant
	invitations      map[string]ports.InventoryAccessInvitation
	customAssetTypes map[customfield.AssetTypeID]customfield.AssetType
	customFields     map[customfield.ID]customfield.Definition
	assets           map[asset.ID]asset.Asset
	undoables        map[string]ports.UndoableOperation
	attachments      map[media.ID]media.Attachment
	providerProfiles map[agentmodel.ProviderProfileID]agentmodel.ProviderProfile
	voiceConfigs     map[tenant.ID]ports.VoiceProviderConfigurationRecord
	providerCreds    map[string]ports.ProviderCredentialRecord
	realtimeSessions map[string]ports.RealtimeSessionRecord
	actionPlans      map[string]ports.ActionPlanRecord
	importJobs       map[string]importjob.Record
	importJobSources map[string]ports.ImportJobSourceRecord
	importLinks      map[string]ports.ImportSourceLink
	importResources  map[string]ports.ImportJobResource
	blobs            map[media.StorageKey][]byte
	blobDeletions    map[string]ports.BlobDeletionEvent
	auditRecords     map[audit.ID]audit.Record
	outbox           map[string]ports.AuthorizationOutboxEvent
}

func NewStore() *Store {
	return &Store{
		users:            map[identity.PrincipalID]identity.User{},
		tenants:          map[tenant.ID]tenant.Tenant{},
		inventories:      map[inventory.InventoryID]inventory.Inventory{},
		accessGrants:     map[string]ports.InventoryAccessGrant{},
		invitations:      map[string]ports.InventoryAccessInvitation{},
		customAssetTypes: map[customfield.AssetTypeID]customfield.AssetType{},
		customFields:     map[customfield.ID]customfield.Definition{},
		assets:           map[asset.ID]asset.Asset{},
		undoables:        map[string]ports.UndoableOperation{},
		attachments:      map[media.ID]media.Attachment{},
		providerProfiles: map[agentmodel.ProviderProfileID]agentmodel.ProviderProfile{},
		voiceConfigs:     map[tenant.ID]ports.VoiceProviderConfigurationRecord{},
		providerCreds:    map[string]ports.ProviderCredentialRecord{},
		realtimeSessions: map[string]ports.RealtimeSessionRecord{},
		actionPlans:      map[string]ports.ActionPlanRecord{},
		importJobs:       map[string]importjob.Record{},
		importJobSources: map[string]ports.ImportJobSourceRecord{},
		importLinks:      map[string]ports.ImportSourceLink{},
		importResources:  map[string]ports.ImportJobResource{},
		blobs:            map[media.StorageKey][]byte{},
		blobDeletions:    map[string]ports.BlobDeletionEvent{},
		auditRecords:     map[audit.ID]audit.Record{},
		outbox:           map[string]ports.AuthorizationOutboxEvent{},
	}
}
