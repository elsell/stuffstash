package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type GrantInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          GrantBody
}

type GrantBody struct {
	PrincipalID  string `json:"principalId" doc:"User principal ID to grant access to"`
	Relationship string `json:"relationship" enum:"viewer,editor" doc:"Direct inventory relationship"`
}

type GrantInventoryAccessOutput struct {
	Body shared.SuccessEnvelope[GrantResponse]
}

type ListInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoryAccessOutput struct {
	Body shared.SuccessEnvelope[[]GrantResponse]
}

type GrantResponse struct {
	TenantID     string `json:"tenantId"`
	InventoryID  string `json:"inventoryId"`
	PrincipalID  string `json:"principalId"`
	Relationship string `json:"relationship"`
}
