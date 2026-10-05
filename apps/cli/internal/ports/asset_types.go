package ports

import "context"

type TypeAction string

const (
	CreateType  TypeAction = "create"
	UpdateType  TypeAction = "update"
	ArchiveType TypeAction = "archive"
	RestoreType TypeAction = "restore"
)

type AssetType struct {
	ID                string  `json:"id"`
	TenantID          string  `json:"tenantId"`
	InventoryID       *string `json:"inventoryId,omitempty"`
	Scope             string  `json:"scope"`
	Key               string  `json:"key"`
	DisplayName       string  `json:"displayName"`
	Description       string  `json:"description"`
	ExpirationEnabled bool    `json:"expirationEnabled"`
	Lifecycle         string  `json:"lifecycleState"`
}
type AssetTypesAPI interface {
	AssetTypes(context.Context, DefinitionScope, Page, string) (Result[[]AssetType], error)
	AssetType(context.Context, DefinitionScope, string) (Result[AssetType], error)
	ChangeAssetType(context.Context, DefinitionScope, string, TypeAction, []byte) (Result[AssetType], error)
	DeleteAssetType(context.Context, DefinitionScope, string) error
}
