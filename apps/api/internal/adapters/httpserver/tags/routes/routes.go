package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tags/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tags/mapper"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/tags", func(ctx context.Context, input *dto.ListAssetTagsInput) (*dto.ListAssetTagsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ListAssetTags(ctx, app.ListAssetTagsInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ListAssetTagsOutput{Body: shared.SuccessEnvelope[[]dto.AssetTagResponse]{
			Data: mapper.AssetTagsToResponse(result.Items),
			Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
		}}, nil
	}, huma.OperationTags("tags"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/tags", func(ctx context.Context, input *dto.CreateAssetTagInput) (*dto.CreateAssetTagOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		tag, err := application.CreateAssetTag(ctx, app.CreateAssetTagInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Key:         input.Body.Key,
			DisplayName: input.Body.DisplayName,
			Color:       input.Body.Color,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.CreateAssetTagOutput{Body: shared.SuccessEnvelope[dto.AssetTagResponse]{
			Data: mapper.AssetTagToResponse(tag),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("tags"), shared.CreatedOperation, shared.SecuredOperation)

	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/tags/{tagId}", func(ctx context.Context, input *dto.UpdateAssetTagInput) (*dto.UpdateAssetTagOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		tag, err := application.UpdateAssetTag(ctx, app.UpdateAssetTagInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			TagID:       assettag.ID(input.TagID),
			DisplayName: input.Body.DisplayName,
			Color:       input.Body.Color,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateAssetTagOutput{Body: shared.SuccessEnvelope[dto.AssetTagResponse]{
			Data: mapper.AssetTagToResponse(tag),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("tags"), shared.SecuredOperation)

	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/tags/{tagId}", func(ctx context.Context, input *dto.ArchiveAssetTagInput) (*dto.ArchiveAssetTagOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		tag, err := application.ArchiveAssetTag(ctx, app.AssetTagLifecycleInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			TagID:       assettag.ID(input.TagID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ArchiveAssetTagOutput{Body: shared.SuccessEnvelope[dto.AssetTagResponse]{
			Data: mapper.AssetTagToResponse(tag),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("tags"), shared.SecuredOperation)
}
