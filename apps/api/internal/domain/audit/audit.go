package audit

import (
	"strings"
	"time"
)

type ID string

func NewID(value string) (ID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return ID(value), true
}

func (id ID) String() string {
	return string(id)
}

type TenantID string

func (id TenantID) String() string {
	return string(id)
}

type InventoryID string

func (id InventoryID) String() string {
	return string(id)
}

type PrincipalID string

func (id PrincipalID) String() string {
	return string(id)
}

type Action string

const (
	ActionTenantCreated                        Action = "tenant.created"
	ActionTenantViewed                         Action = "tenant.viewed"
	ActionTenantListed                         Action = "tenant.listed"
	ActionTenantUpdated                        Action = "tenant.updated"
	ActionTenantArchived                       Action = "tenant.archived"
	ActionTenantRestored                       Action = "tenant.restored"
	ActionTenantDeleted                        Action = "tenant.deleted"
	ActionInventoryCreated                     Action = "inventory.created"
	ActionInventoryViewed                      Action = "inventory.viewed"
	ActionInventoryListed                      Action = "inventory.listed"
	ActionInventoryUpdated                     Action = "inventory.updated"
	ActionInventoryArchived                    Action = "inventory.archived"
	ActionInventoryRestored                    Action = "inventory.restored"
	ActionInventoryDeleted                     Action = "inventory.deleted"
	ActionInventoryAccessGranted               Action = "inventory_access.granted"
	ActionInventoryAccessGrantViewed           Action = "inventory_access_grant.viewed"
	ActionInventoryAccessGrantListed           Action = "inventory_access_grant.listed"
	ActionInventoryAccessRevoked               Action = "inventory_access.revoked"
	ActionInventoryInvitationCreated           Action = "inventory_invitation.created"
	ActionInventoryInvitationViewed            Action = "inventory_invitation.viewed"
	ActionInventoryInvitationListed            Action = "inventory_invitation.listed"
	ActionInventoryInvitationAccepted          Action = "inventory_invitation.accepted"
	ActionInventoryInvitationExpirationUpdated Action = "inventory_invitation.expiration_updated"
	ActionInventoryInvitationRevoked           Action = "inventory_invitation.revoked"
	ActionInventoryInvitationCancelled         Action = "inventory_invitation.cancelled"
	ActionInventoryInvitationDeleted           Action = "inventory_invitation.deleted"
	ActionCustomAssetTypeCreated               Action = "custom_asset_type.created"
	ActionCustomAssetTypeViewed                Action = "custom_asset_type.viewed"
	ActionCustomAssetTypeListed                Action = "custom_asset_type.listed"
	ActionCustomAssetTypeUpdated               Action = "custom_asset_type.updated"
	ActionCustomAssetTypeArchived              Action = "custom_asset_type.archived"
	ActionCustomAssetTypeRestored              Action = "custom_asset_type.restored"
	ActionCustomAssetTypeDeleted               Action = "custom_asset_type.deleted"
	ActionCustomFieldDefinitionCreated         Action = "custom_field_definition.created"
	ActionCustomFieldDefinitionViewed          Action = "custom_field_definition.viewed"
	ActionCustomFieldDefinitionListed          Action = "custom_field_definition.listed"
	ActionCustomFieldDefinitionUpdated         Action = "custom_field_definition.updated"
	ActionCustomFieldDefinitionArchived        Action = "custom_field_definition.archived"
	ActionCustomFieldDefinitionRestored        Action = "custom_field_definition.restored"
	ActionCustomFieldDefinitionDeleted         Action = "custom_field_definition.deleted"
	ActionAssetCreated                         Action = "asset.created"
	ActionAssetViewed                          Action = "asset.viewed"
	ActionAssetListed                          Action = "asset.listed"
	ActionAssetUpdated                         Action = "asset.updated"
	ActionAssetMoved                           Action = "asset.moved"
	ActionAssetArchived                        Action = "asset.archived"
	ActionAssetRestored                        Action = "asset.restored"
	ActionAssetDeleted                         Action = "asset.deleted"
	ActionAttachmentCreated                    Action = "attachment.created"
	ActionAttachmentViewed                     Action = "attachment.viewed"
	ActionAttachmentListed                     Action = "attachment.listed"
	ActionAttachmentContentDownloaded          Action = "attachment.content_downloaded"
	ActionAttachmentArchived                   Action = "attachment.archived"
	ActionAttachmentRestored                   Action = "attachment.restored"
	ActionAttachmentDeleted                    Action = "attachment.deleted"
	ActionAuditRecordListed                    Action = "audit_record.listed"
	ActionUndoableOperationUndone              Action = "undoable_operation.undone"
	ActionUndoableOperationRedone              Action = "undoable_operation.redone"
	ActionProviderProfileCreated               Action = "provider_profile.created"
	ActionProviderProfileViewed                Action = "provider_profile.viewed"
	ActionProviderProfileListed                Action = "provider_profile.listed"
	ActionProviderProfileUpdated               Action = "provider_profile.updated"
	ActionProviderProfileEnabled               Action = "provider_profile.enabled"
	ActionProviderProfileDisabled              Action = "provider_profile.disabled"
	ActionProviderProfileArchived              Action = "provider_profile.archived"
	ActionProviderProfileCredentialReplaced    Action = "provider_profile.credential_replaced"
	ActionProviderProfileTested                Action = "provider_profile.tested"
	ActionVoiceProviderConfigurationUpdated    Action = "voice_provider_configuration.updated"
)

func NewAction(value string) (Action, bool) {
	action := Action(strings.TrimSpace(value))
	switch action {
	case ActionTenantCreated,
		ActionTenantViewed,
		ActionTenantListed,
		ActionTenantUpdated,
		ActionTenantArchived,
		ActionTenantRestored,
		ActionTenantDeleted,
		ActionInventoryCreated,
		ActionInventoryViewed,
		ActionInventoryListed,
		ActionInventoryUpdated,
		ActionInventoryArchived,
		ActionInventoryRestored,
		ActionInventoryDeleted,
		ActionInventoryAccessGranted,
		ActionInventoryAccessGrantViewed,
		ActionInventoryAccessGrantListed,
		ActionInventoryAccessRevoked,
		ActionInventoryInvitationCreated,
		ActionInventoryInvitationViewed,
		ActionInventoryInvitationListed,
		ActionInventoryInvitationAccepted,
		ActionInventoryInvitationExpirationUpdated,
		ActionInventoryInvitationRevoked,
		ActionInventoryInvitationCancelled,
		ActionInventoryInvitationDeleted,
		ActionCustomAssetTypeCreated,
		ActionCustomAssetTypeViewed,
		ActionCustomAssetTypeListed,
		ActionCustomAssetTypeUpdated,
		ActionCustomAssetTypeArchived,
		ActionCustomAssetTypeRestored,
		ActionCustomAssetTypeDeleted,
		ActionCustomFieldDefinitionCreated,
		ActionCustomFieldDefinitionViewed,
		ActionCustomFieldDefinitionListed,
		ActionCustomFieldDefinitionUpdated,
		ActionCustomFieldDefinitionArchived,
		ActionCustomFieldDefinitionRestored,
		ActionCustomFieldDefinitionDeleted,
		ActionAssetCreated,
		ActionAssetViewed,
		ActionAssetListed,
		ActionAssetUpdated,
		ActionAssetMoved,
		ActionAssetArchived,
		ActionAssetRestored,
		ActionAssetDeleted,
		ActionAttachmentCreated,
		ActionAttachmentViewed,
		ActionAttachmentListed,
		ActionAttachmentContentDownloaded,
		ActionAttachmentArchived,
		ActionAttachmentRestored,
		ActionAttachmentDeleted,
		ActionAuditRecordListed,
		ActionUndoableOperationUndone,
		ActionUndoableOperationRedone,
		ActionProviderProfileCreated,
		ActionProviderProfileViewed,
		ActionProviderProfileListed,
		ActionProviderProfileUpdated,
		ActionProviderProfileEnabled,
		ActionProviderProfileDisabled,
		ActionProviderProfileArchived,
		ActionProviderProfileCredentialReplaced,
		ActionProviderProfileTested,
		ActionVoiceProviderConfigurationUpdated:
		return action, true
	default:
		return "", false
	}
}

func (a Action) String() string {
	return string(a)
}

type Source string

const (
	SourceAPI           Source = "api"
	SourceConversation  Source = "conversation"
	SourceMCP           Source = "mcp"
	SourceImport        Source = "import"
	SourceBackgroundJob Source = "background_job"
	SourceSystem        Source = "system"
)

func NewSource(value string) (Source, bool) {
	source := Source(strings.TrimSpace(value))
	switch source {
	case SourceAPI, SourceConversation, SourceMCP, SourceImport, SourceBackgroundJob, SourceSystem:
		return source, true
	default:
		return "", false
	}
}

func (s Source) String() string {
	return string(s)
}

type TargetType string

const (
	TargetTenant                TargetType = "tenant"
	TargetInventory             TargetType = "inventory"
	TargetInventoryAccessGrant  TargetType = "inventory_access_grant"
	TargetInventoryInvitation   TargetType = "inventory_invitation"
	TargetCustomAssetType       TargetType = "custom_asset_type"
	TargetCustomFieldDefinition TargetType = "custom_field_definition"
	TargetAsset                 TargetType = "asset"
	TargetAttachment            TargetType = "attachment"
	TargetAuditRecord           TargetType = "audit_record"
	TargetUndoableOperation     TargetType = "undoable_operation"
	TargetProviderProfile       TargetType = "provider_profile"
)

func NewTargetType(value string) (TargetType, bool) {
	targetType := TargetType(strings.TrimSpace(value))
	switch targetType {
	case TargetTenant, TargetInventory, TargetInventoryAccessGrant, TargetInventoryInvitation, TargetCustomAssetType, TargetCustomFieldDefinition, TargetAsset, TargetAttachment, TargetAuditRecord, TargetUndoableOperation, TargetProviderProfile:
		return targetType, true
	default:
		return "", false
	}
}

func (t TargetType) String() string {
	return string(t)
}

type Record struct {
	ID          ID
	TenantID    TenantID
	InventoryID InventoryID
	PrincipalID PrincipalID
	Action      Action
	Source      Source
	TargetType  TargetType
	TargetID    string
	OccurredAt  time.Time
	RequestID   string
	Metadata    map[string]string
}

func NewRecord(
	id ID,
	tenantID TenantID,
	inventoryID InventoryID,
	principalID PrincipalID,
	action Action,
	source Source,
	targetType TargetType,
	targetID string,
	occurredAt time.Time,
	requestID string,
	metadata map[string]string,
) (Record, bool) {
	if id.String() == "" || tenantID.String() == "" || principalID.String() == "" || action.String() == "" || source.String() == "" || targetType.String() == "" || strings.TrimSpace(targetID) == "" || occurredAt.IsZero() {
		return Record{}, false
	}
	copiedMetadata := map[string]string{}
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		if key == "" {
			return Record{}, false
		}
		copiedMetadata[key] = value
	}
	return Record{
		ID:          id,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		PrincipalID: principalID,
		Action:      action,
		Source:      source,
		TargetType:  targetType,
		TargetID:    strings.TrimSpace(targetID),
		OccurredAt:  occurredAt,
		RequestID:   strings.TrimSpace(requestID),
		Metadata:    copiedMetadata,
	}, true
}

func (r Record) CursorKey() string {
	return r.OccurredAt.UTC().Format(time.RFC3339Nano) + ":" + r.ID.String()
}

func (r Record) Before(other Record) bool {
	if !r.OccurredAt.Equal(other.OccurredAt) {
		return r.OccurredAt.Before(other.OccurredAt)
	}
	return r.ID.String() < other.ID.String()
}

func (r Record) MetadataValues() map[string]string {
	copied := map[string]string{}
	for key, value := range r.Metadata {
		copied[key] = value
	}
	return copied
}
