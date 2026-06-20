package dto

import (
	assetsdto "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type ApplyUndoableOperationInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	OperationID   string `path:"operationId" doc:"Undoable operation ID"`
}

type ApplyUndoableOperationOutput struct {
	Body shared.SuccessEnvelope[assetsdto.AssetResponse]
}
