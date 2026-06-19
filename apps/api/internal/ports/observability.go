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
	EventInventoryCreated                EventName = "inventory.created"
	EventInventoriesListed               EventName = "inventory.listed"
	EventInventoryAccessGranted          EventName = "inventory_access.granted"
	EventInventoryAccessListed           EventName = "inventory_access.listed"
	EventCustomAssetTypeCreated          EventName = "custom_asset_type.created"
	EventCustomAssetTypeUpdated          EventName = "custom_asset_type.updated"
	EventCustomAssetTypesListed          EventName = "custom_asset_type.listed"
	EventCustomFieldDefinitionCreated    EventName = "custom_field_definition.created"
	EventCustomFieldDefinitionUpdated    EventName = "custom_field_definition.updated"
	EventCustomFieldDefinitionsListed    EventName = "custom_field_definition.listed"
	EventAssetCreated                    EventName = "asset.created"
	EventAssetUpdated                    EventName = "asset.updated"
	EventAssetArchived                   EventName = "asset.archived"
	EventAssetRestored                   EventName = "asset.restored"
	EventAssetsListed                    EventName = "asset.listed"
	EventAttachmentCreated               EventName = "attachment.created"
	EventAttachmentsListed               EventName = "attachment.listed"
	EventAttachmentContentDownloaded     EventName = "attachment_content.downloaded"
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
