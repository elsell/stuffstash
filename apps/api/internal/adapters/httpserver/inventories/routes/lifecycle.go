package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterArchive(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/archive", func(ctx context.Context, input *dto.UpdateInventoryLifecycleInput) (*dto.UpdateInventoryLifecycleOutput, error) {
		return updateInventoryLifecycle(ctx, application, input, application.ArchiveInventory)
	}, huma.OperationTags("inventories"), shared.SecuredOperation)
}

func RegisterRestore(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/restore", func(ctx context.Context, input *dto.UpdateInventoryLifecycleInput) (*dto.UpdateInventoryLifecycleOutput, error) {
		return updateInventoryLifecycle(ctx, application, input, application.RestoreInventory)
	}, huma.OperationTags("inventories"), shared.SecuredOperation)
}

func updateInventoryLifecycle(ctx context.Context, application app.App, input *dto.UpdateInventoryLifecycleInput, operation func(context.Context, app.UpdateInventoryLifecycleInput) (inventory.Inventory, error)) (*dto.UpdateInventoryLifecycleOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}
	item, err := operation(ctx, app.UpdateInventoryLifecycleInput{
		Principal:   principal,
		Source:      audit.SourceAPI,
		RequestID:   input.RequestID,
		TenantID:    tenant.ID(input.TenantID),
		InventoryID: inventory.InventoryID(input.InventoryID),
	})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	response, err := inventoryResponse(ctx, application, principal, item)
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.UpdateInventoryLifecycleOutput{Body: shared.SuccessEnvelope[dto.InventoryResponse]{
		Data: response,
		Meta: shared.Meta{TenantID: input.TenantID},
	}}, nil
}

func RegisterDelete(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}", func(ctx context.Context, input *dto.UpdateInventoryLifecycleInput) (*dto.DeleteInventoryOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.DeleteInventory(ctx, app.UpdateInventoryLifecycleInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteInventoryOutput{}, nil
	}, huma.OperationTags("inventories"), shared.NoContentOperation, shared.SecuredOperation)
}
