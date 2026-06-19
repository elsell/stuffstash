package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateTenantDefinitionInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          CreateDefinitionBody
}

type CreateInventoryDefinitionInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          CreateDefinitionBody
}

type CreateDefinitionBody struct {
	Key                string   `json:"key" maxLength:"80" doc:"Stable custom field key"`
	DisplayName        string   `json:"displayName" maxLength:"120" doc:"User-facing field label"`
	Type               string   `json:"type" enum:"text,number,boolean,date,url,enum" doc:"Custom field type"`
	EnumOptions        []string `json:"enumOptions,omitempty" doc:"Allowed enum option keys"`
	Applicability      string   `json:"applicability,omitempty" enum:"all_assets,custom_asset_types" doc:"Assets this field applies to"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds,omitempty" doc:"Custom asset type IDs this field targets"`
}

type CreateDefinitionOutput struct {
	Body shared.SuccessEnvelope[DefinitionResponse]
}

type UpdateTenantDefinitionInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	DefinitionID  string `path:"definitionId" doc:"Custom field definition ID"`
	Body          UpdateDefinitionBody
}

type UpdateInventoryDefinitionInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	DefinitionID  string `path:"definitionId" doc:"Custom field definition ID"`
	Body          UpdateDefinitionBody
}

type UpdateDefinitionBody struct {
	DisplayName        *string  `json:"displayName,omitempty" maxLength:"120" doc:"User-facing field label"`
	Key                *string  `json:"key,omitempty" doc:"Immutable field key; rejected on update"`
	Type               *string  `json:"type,omitempty" doc:"Immutable field type; rejected on update"`
	EnumOptions        []string `json:"enumOptions,omitempty" doc:"Immutable enum options; rejected on update"`
	Applicability      *string  `json:"applicability,omitempty" doc:"Immutable applicability; rejected on update"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds,omitempty" doc:"Immutable custom asset type targets; rejected on update"`
}

func (b UpdateDefinitionBody) HasImmutableFields() bool {
	return b.Key != nil || b.Type != nil || len(b.EnumOptions) != 0 || b.Applicability != nil || len(b.CustomAssetTypeIDs) != 0
}

type UpdateDefinitionOutput struct {
	Body shared.SuccessEnvelope[DefinitionResponse]
}

type ListTenantDefinitionsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoryDefinitionsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListDefinitionsOutput struct {
	Body shared.SuccessEnvelope[[]DefinitionResponse]
}

type DefinitionResponse struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenantId"`
	InventoryID        string   `json:"inventoryId,omitempty"`
	Scope              string   `json:"scope"`
	Key                string   `json:"key"`
	DisplayName        string   `json:"displayName"`
	Type               string   `json:"type"`
	EnumOptions        []string `json:"enumOptions"`
	Applicability      string   `json:"applicability"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds"`
}
