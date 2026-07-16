package dto

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type ListAssetActivityInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	View          string `query:"view" enum:"changes,all" default:"changes" doc:"Meaningful changes or all technical events"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListAssetActivityOutput struct {
	Body shared.SuccessEnvelope[[]AssetActivityResponse]
}

type AssetActivityResponse struct {
	ID          string                        `json:"id"`
	PrincipalID string                        `json:"principalId"`
	Principal   *AuditPrincipalResponse       `json:"principal,omitempty"`
	Action      string                        `json:"action"`
	Category    string                        `json:"category"`
	Source      string                        `json:"source"`
	OccurredAt  time.Time                     `json:"occurredAt"`
	RequestID   string                        `json:"requestId,omitempty"`
	Changes     []AssetActivityChangeResponse `json:"changes"`
	Undo        *AssetActivityUndoResponse    `json:"undo,omitempty"`
	Technical   map[string]string             `json:"technical"`
}

type AssetActivityChangeResponse struct {
	Field         string `json:"field"`
	PreviousValue string `json:"previousValue,omitempty"`
	CurrentValue  string `json:"currentValue,omitempty"`
}

type AssetActivityUndoResponse struct {
	OperationID string `json:"operationId"`
	Status      string `json:"status"`
}
