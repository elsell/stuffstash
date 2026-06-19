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

type ListInventoriesInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoriesOutput struct {
	Body shared.SuccessEnvelope[[]InventoryResponse]
}

type InventoryResponse struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
}
