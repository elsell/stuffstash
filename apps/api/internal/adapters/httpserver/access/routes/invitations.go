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
		response := mapper.InvitationToResponse(result.Invitation)
		response.AcceptanceToken = result.AcceptanceToken
		return &dto.CreateInventoryAccessInvitationOutput{
			Body: shared.SuccessEnvelope[dto.InvitationResponse]{
				Data: response,
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.CreatedOperation, shared.SecuredOperation)

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
