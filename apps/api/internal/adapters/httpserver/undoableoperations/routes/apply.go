package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	assetsdto "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	assetsmapper "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/undoableoperations/dto"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	RegisterUndo(api, application)
	RegisterRedo(api, application)
}

func RegisterUndo(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/undo", func(ctx context.Context, input *dto.ApplyUndoableOperationInput) (*dto.ApplyUndoableOperationOutput, error) {
		return apply(ctx, application, input, application.UndoOperation)
	}, huma.OperationTags("undoable-operations"), shared.SecuredOperation)
}

func RegisterRedo(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/redo", func(ctx context.Context, input *dto.ApplyUndoableOperationInput) (*dto.ApplyUndoableOperationOutput, error) {
		return apply(ctx, application, input, application.RedoOperation)
	}, huma.OperationTags("undoable-operations"), shared.SecuredOperation)
}

func apply(ctx context.Context, application app.App, input *dto.ApplyUndoableOperationInput, applyOperation func(context.Context, app.ApplyUndoableOperationInput) (asset.Asset, error)) (*dto.ApplyUndoableOperationOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}

	item, err := applyOperation(ctx, app.ApplyUndoableOperationInput{
		Principal:   principal,
		Source:      audit.SourceAPI,
		RequestID:   input.RequestID,
		TenantID:    tenant.ID(input.TenantID),
		InventoryID: inventory.InventoryID(input.InventoryID),
		OperationID: input.OperationID,
	})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}

	return &dto.ApplyUndoableOperationOutput{
		Body: shared.SuccessEnvelope[assetsdto.AssetResponse]{
			Data: assetsmapper.AssetToResponse(item, nil, nil),
			Meta: shared.Meta{TenantID: input.TenantID},
		},
	}, nil
}
