package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	registerActivation(api, application)
	huma.Post(api, "/tenants/{tenantId}/conversation-workflows", func(ctx context.Context, input *dto.CreateInput) (*dto.RevisionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		revision, err := application.SaveConversationWorkflowRevision(ctx, app.SaveConversationWorkflowInput{Principal: principal, TenantID: tenant.ID(input.TenantID), Definition: mapper.DefinitionToDomain(input.Body.Definition), Source: audit.SourceAPI, RequestID: input.RequestID})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return revisionOutput(revision, input.TenantID), nil
	}, huma.OperationTags("conversation workflows"), shared.CreatedOperation, shared.SecuredOperation)
	huma.Post(api, "/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions", func(ctx context.Context, input *dto.AppendInput) (*dto.RevisionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		revision, err := application.SaveConversationWorkflowRevision(ctx, app.SaveConversationWorkflowInput{Principal: principal, TenantID: tenant.ID(input.TenantID), WorkflowID: agentmodel.WorkflowID(input.WorkflowID), ExpectedRevision: input.Body.ExpectedRevision, Definition: mapper.DefinitionToDomain(input.Body.Definition), Source: audit.SourceAPI, RequestID: input.RequestID})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return revisionOutput(revision, input.TenantID), nil
	}, huma.OperationTags("conversation workflows"), shared.CreatedOperation, shared.SecuredOperation)
}
func revisionOutput(value agentmodel.WorkflowRevision, tenantID string) *dto.RevisionOutput {
	return &dto.RevisionOutput{Body: shared.SuccessEnvelope[dto.Revision]{Data: mapper.RevisionToResponse(value), Meta: shared.Meta{TenantID: tenantID}}}
}
