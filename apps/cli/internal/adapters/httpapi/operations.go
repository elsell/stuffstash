package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) ApplyOperation(ctx context.Context, s ports.Scope, id string, action ports.OperationAction) (ports.Result[ports.Asset], error) {
	switch action {
	case ports.UndoOperation:
		return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdUndoableOperationsByOperationIdUndo(ctx, s.Tenant, s.Inventory, id, nil)))
	case ports.RedoOperation:
		return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdUndoableOperationsByOperationIdRedo(ctx, s.Tenant, s.Inventory, id, nil)))
	default:
		return ports.Result[ports.Asset]{}, ports.Failure("usage", "Use operations undo or operations redo.")
	}
}
