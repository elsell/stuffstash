package ports

import "context"

type EventName string

const (
	EventHealthChecked                   EventName = "health.checked"
	EventHTTPServerStartFailed           EventName = "http.server.start_failed"
	EventHTTPServerShutdownFailed        EventName = "http.server.shutdown_failed"
	EventApplicationStartupFailed        EventName = "application.startup_failed"
	EventApplicationShutdownFailed       EventName = "application.shutdown_failed"
	EventAuthenticationFailed            EventName = "authentication.failed"
	EventAuthorizationDenied             EventName = "authorization.denied"
	EventTenantCreated                   EventName = "tenant.created"
	EventTenantViewed                    EventName = "tenant.viewed"
	EventTenantUpdated                   EventName = "tenant.updated"
	EventTenantArchived                  EventName = "tenant.archived"
	EventTenantRestored                  EventName = "tenant.restored"
	EventTenantDeleted                   EventName = "tenant.deleted"
	EventInventoryCreated                EventName = "inventory.created"
	EventInventoryViewed                 EventName = "inventory.viewed"
	EventInventoryUpdated                EventName = "inventory.updated"
	EventInventoryArchived               EventName = "inventory.archived"
	EventInventoryRestored               EventName = "inventory.restored"
	EventInventoryDeleted                EventName = "inventory.deleted"
	EventInventoriesListed               EventName = "inventory.listed"
	EventInventoryAccessGranted          EventName = "inventory_access.granted"
	EventInventoryAccessViewed           EventName = "inventory_access.viewed"
	EventInventoryAccessRevoked          EventName = "inventory_access.revoked"
	EventInventoryAccessListed           EventName = "inventory_access.listed"
	EventInventoryInvitationCreated      EventName = "inventory_invitation.created"
	EventInventoryInvitationViewed       EventName = "inventory_invitation.viewed"
	EventInventoryInvitationAccepted     EventName = "inventory_invitation.accepted"
	EventInventoryInvitationRevoked      EventName = "inventory_invitation.revoked"
	EventInventoryInvitationCancelled    EventName = "inventory_invitation.cancelled"
	EventInventoryInvitationDeleted      EventName = "inventory_invitation.deleted"
	EventCustomAssetTypeCreated          EventName = "custom_asset_type.created"
	EventCustomAssetTypeViewed           EventName = "custom_asset_type.viewed"
	EventCustomAssetTypeUpdated          EventName = "custom_asset_type.updated"
	EventCustomAssetTypeArchived         EventName = "custom_asset_type.archived"
	EventCustomAssetTypeRestored         EventName = "custom_asset_type.restored"
	EventCustomAssetTypeDeleted          EventName = "custom_asset_type.deleted"
	EventCustomAssetTypesListed          EventName = "custom_asset_type.listed"
	EventCustomFieldDefinitionCreated    EventName = "custom_field_definition.created"
	EventCustomFieldDefinitionViewed     EventName = "custom_field_definition.viewed"
	EventCustomFieldDefinitionUpdated    EventName = "custom_field_definition.updated"
	EventCustomFieldDefinitionArchived   EventName = "custom_field_definition.archived"
	EventCustomFieldDefinitionRestored   EventName = "custom_field_definition.restored"
	EventCustomFieldDefinitionDeleted    EventName = "custom_field_definition.deleted"
	EventCustomFieldDefinitionsListed    EventName = "custom_field_definition.listed"
	EventAssetCreated                    EventName = "asset.created"
	EventAssetViewed                     EventName = "asset.viewed"
	EventAssetUpdated                    EventName = "asset.updated"
	EventAssetArchived                   EventName = "asset.archived"
	EventAssetRestored                   EventName = "asset.restored"
	EventAssetDeleted                    EventName = "asset.deleted"
	EventAssetsListed                    EventName = "asset.listed"
	EventAssetsSearched                  EventName = "asset.searched"
	EventAttachmentCreated               EventName = "attachment.created"
	EventAttachmentViewed                EventName = "attachment.viewed"
	EventAttachmentsListed               EventName = "attachment.listed"
	EventAttachmentContentDownloaded     EventName = "attachment_content.downloaded"
	EventAttachmentArchived              EventName = "attachment.archived"
	EventAttachmentRestored              EventName = "attachment.restored"
	EventAttachmentDeleted               EventName = "attachment.deleted"
	EventBlobStorageFailed               EventName = "blob_storage.failed"
	EventAuditRecordsListed              EventName = "audit_record.listed"
	EventAuthorizationOutboxDrained      EventName = "authorization_outbox.drained"
	EventAuthorizationOutboxFailed       EventName = "authorization_outbox.failed"
	EventAuthorizationOutboxDeadLettered EventName = "authorization_outbox.dead_lettered"
)

type Event struct {
	Name    EventName
	Message string
	Fields  map[string]string
}

type Observer interface {
	Record(ctx context.Context, event Event)
}
