package ports

import "context"

type FieldAction string

const (
	CreateField  FieldAction = "create"
	UpdateField  FieldAction = "update"
	ArchiveField FieldAction = "archive"
	RestoreField FieldAction = "restore"
)

type FieldDefinition struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenantId"`
	InventoryID        *string  `json:"inventoryId,omitempty"`
	Scope              string   `json:"scope"`
	Key                string   `json:"key"`
	DisplayName        string   `json:"displayName"`
	Type               string   `json:"type"`
	EnumOptions        []string `json:"enumOptions"`
	Applicability      string   `json:"applicability"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds"`
	Lifecycle          string   `json:"lifecycleState"`
}
type FieldDefinitionsAPI interface {
	FieldDefinitions(context.Context, DefinitionScope, Page, string) (Result[[]FieldDefinition], error)
	FieldDefinition(context.Context, DefinitionScope, string) (Result[FieldDefinition], error)
	ChangeFieldDefinition(context.Context, DefinitionScope, string, FieldAction, []byte) (Result[FieldDefinition], error)
	DeleteFieldDefinition(context.Context, DefinitionScope, string) error
}
