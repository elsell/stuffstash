package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterArchiveTenant(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/custom-field-definitions/{definitionId}/archive", func(ctx context.Context, input *dto.GetTenantDefinitionInput) (*dto.UpdateDefinitionLifecycleOutput, error) {
		return updateTenantDefinitionLifecycle(ctx, application, input, application.ArchiveTenantCustomFieldDefinition)
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}

func RegisterRestoreTenant(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/custom-field-definitions/{definitionId}/restore", func(ctx context.Context, input *dto.GetTenantDefinitionInput) (*dto.UpdateDefinitionLifecycleOutput, error) {
		return updateTenantDefinitionLifecycle(ctx, application, input, application.RestoreTenantCustomFieldDefinition)
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}

func updateTenantDefinitionLifecycle(ctx context.Context, application app.App, input *dto.GetTenantDefinitionInput, operation func(context.Context, app.UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error)) (*dto.UpdateDefinitionLifecycleOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}
	definition, err := operation(ctx, app.UpdateCustomFieldDefinitionLifecycleInput{
		Principal:    principal,
		Source:       audit.SourceAPI,
		RequestID:    input.RequestID,
		TenantID:     tenant.ID(input.TenantID),
		DefinitionID: customfield.ID(input.DefinitionID),
	})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.UpdateDefinitionLifecycleOutput{Body: shared.SuccessEnvelope[dto.DefinitionResponse]{
		Data: mapper.DefinitionToResponse(definition),
		Meta: shared.Meta{TenantID: input.TenantID},
	}}, nil
}

func RegisterArchiveInventory(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/archive", func(ctx context.Context, input *dto.GetInventoryDefinitionInput) (*dto.UpdateDefinitionLifecycleOutput, error) {
		return updateInventoryDefinitionLifecycle(ctx, application, input, application.ArchiveInventoryCustomFieldDefinition)
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}

func RegisterRestoreInventory(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/restore", func(ctx context.Context, input *dto.GetInventoryDefinitionInput) (*dto.UpdateDefinitionLifecycleOutput, error) {
		return updateInventoryDefinitionLifecycle(ctx, application, input, application.RestoreInventoryCustomFieldDefinition)
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}

func updateInventoryDefinitionLifecycle(ctx context.Context, application app.App, input *dto.GetInventoryDefinitionInput, operation func(context.Context, app.UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error)) (*dto.UpdateDefinitionLifecycleOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}
	definition, err := operation(ctx, app.UpdateCustomFieldDefinitionLifecycleInput{
		Principal:    principal,
		Source:       audit.SourceAPI,
		RequestID:    input.RequestID,
		TenantID:     tenant.ID(input.TenantID),
		InventoryID:  inventory.InventoryID(input.InventoryID),
		DefinitionID: customfield.ID(input.DefinitionID),
	})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.UpdateDefinitionLifecycleOutput{Body: shared.SuccessEnvelope[dto.DefinitionResponse]{
		Data: mapper.DefinitionToResponse(definition),
		Meta: shared.Meta{TenantID: input.TenantID},
	}}, nil
}

func RegisterDeleteTenant(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/custom-field-definitions/{definitionId}", func(ctx context.Context, input *dto.GetTenantDefinitionInput) (*dto.DeleteDefinitionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.DeleteTenantCustomFieldDefinition(ctx, app.UpdateCustomFieldDefinitionLifecycleInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			DefinitionID: customfield.ID(input.DefinitionID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteDefinitionOutput{}, nil
	}, huma.OperationTags("custom field definitions"), shared.NoContentOperation, shared.SecuredOperation)
}

func RegisterDeleteInventory(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", func(ctx context.Context, input *dto.GetInventoryDefinitionInput) (*dto.DeleteDefinitionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.DeleteInventoryCustomFieldDefinition(ctx, app.UpdateCustomFieldDefinitionLifecycleInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			DefinitionID: customfield.ID(input.DefinitionID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteDefinitionOutput{}, nil
	}, huma.OperationTags("custom field definitions"), shared.NoContentOperation, shared.SecuredOperation)
}
