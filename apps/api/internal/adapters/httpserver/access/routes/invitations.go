package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterInvitations(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", func(ctx context.Context, input *dto.CreateInventoryAccessInvitationInput) (*dto.CreateInventoryAccessInvitationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.CreateInventoryAccessInvitation(ctx, app.CreateInventoryAccessInvitationInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			Email:        input.Body.Email,
			Relationship: input.Body.Relationship,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.CreateInventoryAccessInvitationOutput{
			CacheControl: "no-store",
			Body: shared.SuccessEnvelope[dto.CreatedInvitationResponse]{
				Data: mapper.CreatedInvitationToResponse(result.Invitation, result.InviteURL),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.CreatedOperation, shared.SecuredOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", func(ctx context.Context, input *dto.ListInventoryAccessInvitationsInput) (*dto.ListInventoryAccessInvitationsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventoryAccessInvitations(ctx, app.ListInventoryAccessInvitationsInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			Limit:        input.Limit,
			Cursor:       input.Cursor,
			StatusFilter: input.Status,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListInventoryAccessInvitationsOutput{
			Body: shared.SuccessEnvelope[[]dto.InvitationResponse]{
				Data: mapper.InvitationsToResponseAt(result.Items, result.Now),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/preview", func(ctx context.Context, input *dto.PreviewInventoryAccessInvitationInput) (*dto.PreviewInventoryAccessInvitationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		preview, err := application.PreviewInventoryAccessInvitation(ctx, app.PreviewInventoryAccessInvitationInput{
			Principal:    principal,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			InvitationID: input.InvitationID,
			Token:        input.Body.AcceptanceToken,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PreviewInventoryAccessInvitationOutput{
			Body: shared.SuccessEnvelope[dto.InvitationPreviewResponse]{
				Data: mapper.InvitationPreviewToResponse(
					preview.InventoryID,
					preview.InventoryName,
					preview.Relationship,
					preview.Status,
					preview.ExpiresAt,
					preview.IsExpired,
				),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept", func(ctx context.Context, input *dto.AcceptInventoryAccessInvitationInput) (*dto.AcceptInventoryAccessInvitationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		invitation, grant, err := application.AcceptInventoryAccessInvitation(ctx, app.AcceptInventoryAccessInvitationInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			InvitationID: input.InvitationID,
			Token:        input.Body.AcceptanceToken,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.AcceptInventoryAccessInvitationOutput{
			Body: shared.SuccessEnvelope[dto.InvitationAcceptanceResponse]{
				Data: mapper.InvitationAcceptanceToResponse(invitation, grant),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.SecuredOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", func(ctx context.Context, input *dto.RevokeInventoryAccessInvitationInput) (*dto.GetInventoryAccessInvitationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		invitation, err := application.GetInventoryAccessInvitation(ctx, app.GetInventoryAccessInvitationInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			InvitationID: input.InvitationID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetInventoryAccessInvitationOutput{
			Body: shared.SuccessEnvelope[dto.InvitationResponse]{
				Data: mapper.InvitationToResponse(invitation),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.SecuredOperation)

	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/expiration", func(ctx context.Context, input *dto.UpdateInventoryAccessInvitationExpirationInput) (*dto.UpdateInventoryAccessInvitationExpirationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		invitation, err := application.UpdateInventoryAccessInvitationExpiration(ctx, app.UpdateInventoryAccessInvitationExpirationInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			InvitationID: input.InvitationID,
			ExpiresAt:    input.Body.ExpiresAt,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateInventoryAccessInvitationExpirationOutput{
			Body: shared.SuccessEnvelope[dto.InvitationResponse]{
				Data: mapper.InvitationToResponse(invitation),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.SecuredOperation)

	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel", func(ctx context.Context, input *dto.RevokeInventoryAccessInvitationInput) (*dto.RevokeInventoryAccessInvitationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if _, err := application.CancelInventoryAccessInvitation(ctx, app.RevokeInventoryAccessInvitationInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			InvitationID: input.InvitationID,
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.RevokeInventoryAccessInvitationOutput{}, nil
	}, huma.OperationTags("inventory access"), shared.NoContentOperation, shared.SecuredOperation)

	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", func(ctx context.Context, input *dto.RevokeInventoryAccessInvitationInput) (*dto.RevokeInventoryAccessInvitationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if _, err := application.DeleteInventoryAccessInvitation(ctx, app.RevokeInventoryAccessInvitationInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			InvitationID: input.InvitationID,
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.RevokeInventoryAccessInvitationOutput{}, nil
	}, huma.OperationTags("inventory access"), shared.NoContentOperation, shared.SecuredOperation)
}
