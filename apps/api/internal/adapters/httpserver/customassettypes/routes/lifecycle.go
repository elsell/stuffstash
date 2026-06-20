package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterRestoreTenant(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/restore", func(ctx context.Context, input *dto.ArchiveTenantAssetTypeInput) (*dto.ArchiveAssetTypeOutput, error) {
		return updateTenantAssetTypeLifecycle(ctx, application, input, application.RestoreTenantCustomAssetType)
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}

func RegisterRestoreInventory(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/restore", func(ctx context.Context, input *dto.ArchiveInventoryAssetTypeInput) (*dto.ArchiveAssetTypeOutput, error) {
		return updateInventoryAssetTypeLifecycle(ctx, application, input, application.RestoreInventoryCustomAssetType)
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}

func updateTenantAssetTypeLifecycle(ctx context.Context, application app.App, input *dto.ArchiveTenantAssetTypeInput, operation func(context.Context, app.ArchiveCustomAssetTypeInput) (customfield.AssetType, error)) (*dto.ArchiveAssetTypeOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}
	assetType, err := operation(ctx, app.ArchiveCustomAssetTypeInput{
		Principal:         principal,
		Source:            audit.SourceAPI,
		RequestID:         input.RequestID,
		TenantID:          tenant.ID(input.TenantID),
		CustomAssetTypeID: customfield.AssetTypeID(input.CustomAssetTypeID),
	})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.ArchiveAssetTypeOutput{Body: shared.SuccessEnvelope[dto.AssetTypeResponse]{
		Data: mapper.AssetTypeToResponse(assetType),
		Meta: shared.Meta{TenantID: input.TenantID},
	}}, nil
}

func updateInventoryAssetTypeLifecycle(ctx context.Context, application app.App, input *dto.ArchiveInventoryAssetTypeInput, operation func(context.Context, app.ArchiveCustomAssetTypeInput) (customfield.AssetType, error)) (*dto.ArchiveAssetTypeOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}
	assetType, err := operation(ctx, app.ArchiveCustomAssetTypeInput{
		Principal:         principal,
		Source:            audit.SourceAPI,
		RequestID:         input.RequestID,
		TenantID:          tenant.ID(input.TenantID),
		InventoryID:       inventory.InventoryID(input.InventoryID),
		CustomAssetTypeID: customfield.AssetTypeID(input.CustomAssetTypeID),
	})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.ArchiveAssetTypeOutput{Body: shared.SuccessEnvelope[dto.AssetTypeResponse]{
		Data: mapper.AssetTypeToResponse(assetType),
		Meta: shared.Meta{TenantID: input.TenantID},
	}}, nil
}

func RegisterDeleteTenant(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", func(ctx context.Context, input *dto.ArchiveTenantAssetTypeInput) (*dto.DeleteAssetTypeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.DeleteTenantCustomAssetType(ctx, app.ArchiveCustomAssetTypeInput{
			Principal:         principal,
			Source:            audit.SourceAPI,
			RequestID:         input.RequestID,
			TenantID:          tenant.ID(input.TenantID),
			CustomAssetTypeID: customfield.AssetTypeID(input.CustomAssetTypeID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteAssetTypeOutput{}, nil
	}, huma.OperationTags("custom asset types"), shared.NoContentOperation, shared.SecuredOperation)
}

func RegisterDeleteInventory(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", func(ctx context.Context, input *dto.ArchiveInventoryAssetTypeInput) (*dto.DeleteAssetTypeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.DeleteInventoryCustomAssetType(ctx, app.ArchiveCustomAssetTypeInput{
			Principal:         principal,
			Source:            audit.SourceAPI,
			RequestID:         input.RequestID,
			TenantID:          tenant.ID(input.TenantID),
			InventoryID:       inventory.InventoryID(input.InventoryID),
			CustomAssetTypeID: customfield.AssetTypeID(input.CustomAssetTypeID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteAssetTypeOutput{}, nil
	}, huma.OperationTags("custom asset types"), shared.NoContentOperation, shared.SecuredOperation)
}
