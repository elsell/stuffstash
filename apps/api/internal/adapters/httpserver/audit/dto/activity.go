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
	ID                string                        `json:"id"`
	PrincipalID       string                        `json:"principalId"`
	Principal         *AuditPrincipalResponse       `json:"principal,omitempty"`
	Action            AssetActivityAction           `json:"action" enum:"asset.created,asset.updated,asset.moved,asset.archived,asset.restored,asset.deleted,asset.checked_out,asset.returned,asset.return_details_updated,asset.viewed,asset.listed,asset.searched,audit_record.listed,undoable_operation.undone,undoable_operation.redone"`
	Category          AssetActivityCategory         `json:"category" enum:"change,read"`
	Source            AssetActivitySource           `json:"source" enum:"api,conversation,mcp,import,background_job,system"`
	OccurredAt        time.Time                     `json:"occurredAt"`
	RequestID         string                        `json:"requestId,omitempty"`
	Changes           []AssetActivityChangeResponse `json:"changes"`
	Undo              *AssetActivityUndoResponse    `json:"undo,omitempty"`
	TechnicalMetadata map[string]string             `json:"technicalMetadata"`
}

type AssetActivityChangeResponse struct {
	Field         AssetActivityField `json:"field" enum:"title,description,tags,parent,lifecycle_state,checkout_state"`
	PreviousValue string             `json:"previousValue,omitempty"`
	CurrentValue  string             `json:"currentValue,omitempty"`
}

type AssetActivityUndoResponse struct {
	OperationID string                  `json:"operationId"`
	Status      AssetActivityUndoStatus `json:"status" enum:"available,undone,redone"`
}

type AssetActivityAction string
type AssetActivityCategory string
type AssetActivitySource string
type AssetActivityField string
type AssetActivityUndoStatus string
