package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateAssetTagInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          CreateAssetTagBody
}

type CreateAssetTagBody struct {
	Key         string `json:"key,omitempty" maxLength:"80" doc:"Stable tag key"`
	DisplayName string `json:"displayName" maxLength:"80" doc:"User-facing tag name"`
	Color       string `json:"color,omitempty" doc:"Optional #RRGGBB tag color"`
}

type CreateAssetTagOutput struct {
	Body shared.SuccessEnvelope[AssetTagResponse]
}

type UpdateAssetTagInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	TagID         string `path:"tagId" doc:"Tag ID"`
	Body          UpdateAssetTagBody
}

type UpdateAssetTagBody struct {
	DisplayName *string `json:"displayName,omitempty" maxLength:"80" doc:"User-facing tag name"`
	Color       *string `json:"color,omitempty" doc:"Optional #RRGGBB tag color"`
}

type UpdateAssetTagOutput struct {
	Body shared.SuccessEnvelope[AssetTagResponse]
}

type ArchiveAssetTagInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	TagID         string `path:"tagId" doc:"Tag ID"`
}

type ArchiveAssetTagOutput struct {
	Body shared.SuccessEnvelope[AssetTagResponse]
}

type ListAssetTagsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"0" doc:"Requested page size; 0 uses the default page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListAssetTagsOutput struct {
	Body shared.SuccessEnvelope[[]AssetTagResponse]
}

type AssetTagResponse struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	InventoryID    string `json:"inventoryId"`
	Key            string `json:"key"`
	DisplayName    string `json:"displayName"`
	Color          string `json:"color,omitempty"`
	LifecycleState string `json:"lifecycleState"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}
