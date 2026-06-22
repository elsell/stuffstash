package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateInventoryInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          CreateInventoryBody
}

type CreateInventoryBody struct {
	Name string `json:"name" maxLength:"120" doc:"Inventory name"`
}

type CreateInventoryOutput struct {
	Body shared.SuccessEnvelope[InventoryResponse]
}

type GetInventoryInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
}

type GetInventoryOutput struct {
	Body shared.SuccessEnvelope[InventoryResponse]
}

type UpdateInventoryInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          UpdateInventoryBody
}

type UpdateInventoryBody struct {
	Name *string `json:"name,omitempty" maxLength:"120" doc:"Inventory name"`
}

type UpdateInventoryOutput struct {
	Body shared.SuccessEnvelope[InventoryResponse]
}

type UpdateInventoryLifecycleInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
}

type UpdateInventoryLifecycleOutput struct {
	Body shared.SuccessEnvelope[InventoryResponse]
}

type DeleteInventoryOutput struct{}

type ListInventoriesInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoriesOutput struct {
	Body shared.SuccessEnvelope[[]InventoryResponse]
}

type InventoryResponse struct {
	ID             string                `json:"id"`
	TenantID       string                `json:"tenantId"`
	Name           string                `json:"name"`
	LifecycleState string                `json:"lifecycleState"`
	Access         shared.AccessResponse `json:"access"`
}
