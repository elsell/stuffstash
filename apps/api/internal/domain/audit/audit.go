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
	ActionTenantCreated                Action = "tenant.created"
	ActionInventoryCreated             Action = "inventory.created"
	ActionInventoryAccessGranted       Action = "inventory_access.granted"
	ActionCustomFieldDefinitionCreated Action = "custom_field_definition.created"
	ActionAssetCreated                 Action = "asset.created"
	ActionAssetUpdated                 Action = "asset.updated"
	ActionAssetMoved                   Action = "asset.moved"
)

func NewAction(value string) (Action, bool) {
	action := Action(strings.TrimSpace(value))
	switch action {
	case ActionTenantCreated,
		ActionInventoryCreated,
		ActionInventoryAccessGranted,
		ActionCustomFieldDefinitionCreated,
		ActionAssetCreated,
		ActionAssetUpdated,
		ActionAssetMoved:
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
	TargetCustomFieldDefinition TargetType = "custom_field_definition"
	TargetAsset                 TargetType = "asset"
)

func NewTargetType(value string) (TargetType, bool) {
	targetType := TargetType(strings.TrimSpace(value))
	switch targetType {
	case TargetTenant, TargetInventory, TargetInventoryAccessGrant, TargetCustomFieldDefinition, TargetAsset:
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
	return r.ID.String()
}

func (r Record) MetadataValues() map[string]string {
	copied := map[string]string{}
	for key, value := range r.Metadata {
		copied[key] = value
	}
	return copied
}
