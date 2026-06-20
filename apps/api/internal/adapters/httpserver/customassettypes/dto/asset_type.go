package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateTenantAssetTypeInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          CreateAssetTypeBody
}

type CreateInventoryAssetTypeInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          CreateAssetTypeBody
}

type CreateAssetTypeBody struct {
	Key         string `json:"key" maxLength:"80" doc:"Stable custom asset type key"`
	DisplayName string `json:"displayName" maxLength:"120" doc:"User-facing custom asset type label"`
	Description string `json:"description,omitempty" maxLength:"1000" doc:"Custom asset type description"`
}

type CreateAssetTypeOutput struct {
	Body shared.SuccessEnvelope[AssetTypeResponse]
}

type UpdateTenantAssetTypeInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	CustomAssetTypeID string `path:"customAssetTypeId" doc:"Custom asset type ID"`
	Body              UpdateAssetTypeBody
}

type UpdateInventoryAssetTypeInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	InventoryID       string `path:"inventoryId" doc:"Inventory ID"`
	CustomAssetTypeID string `path:"customAssetTypeId" doc:"Custom asset type ID"`
	Body              UpdateAssetTypeBody
}

type UpdateAssetTypeBody struct {
	DisplayName *string `json:"displayName,omitempty" maxLength:"120" doc:"User-facing custom asset type label"`
	Description *string `json:"description,omitempty" maxLength:"1000" doc:"Custom asset type description"`
}

type UpdateAssetTypeOutput struct {
	Body shared.SuccessEnvelope[AssetTypeResponse]
}

type ArchiveTenantAssetTypeInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	CustomAssetTypeID string `path:"customAssetTypeId" doc:"Custom asset type ID"`
}

type ArchiveInventoryAssetTypeInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	InventoryID       string `path:"inventoryId" doc:"Inventory ID"`
	CustomAssetTypeID string `path:"customAssetTypeId" doc:"Custom asset type ID"`
}

type ArchiveAssetTypeOutput struct {
	Body shared.SuccessEnvelope[AssetTypeResponse]
}

type ListTenantAssetTypesInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoryAssetTypesInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListAssetTypesOutput struct {
	Body shared.SuccessEnvelope[[]AssetTypeResponse]
}

type AssetTypeResponse struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	InventoryID    string `json:"inventoryId,omitempty"`
	Scope          string `json:"scope"`
	Key            string `json:"key"`
	DisplayName    string `json:"displayName"`
	Description    string `json:"description"`
	LifecycleState string `json:"lifecycleState"`
}
