package dto

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type ListTenantAuditRecordsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoryAuditRecordsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListAssetAuditHistoryInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
}

type ListAuditRecordsOutput struct {
	Body shared.SuccessEnvelope[[]RecordResponse]
}

type RecordResponse struct {
	ID          string                  `json:"id"`
	TenantID    string                  `json:"tenantId"`
	InventoryID string                  `json:"inventoryId,omitempty"`
	PrincipalID string                  `json:"principalId"`
	Principal   *AuditPrincipalResponse `json:"principal,omitempty"`
	Action      string                  `json:"action"`
	Source      string                  `json:"source"`
	TargetType  string                  `json:"targetType"`
	TargetID    string                  `json:"targetId"`
	OccurredAt  time.Time               `json:"occurredAt"`
	RequestID   string                  `json:"requestId,omitempty"`
	Metadata    map[string]string       `json:"metadata"`
}

type AuditPrincipalResponse struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}
