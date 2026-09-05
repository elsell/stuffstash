package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func registerQueries(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/conversation-workflows/{workflowId}", func(ctx context.Context, input *dto.GetInput) (*dto.RevisionOutput, error) {
		return getRevision(ctx, application, *input, "")
	}, huma.OperationTags("conversation workflows"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions/{revisionId}", func(ctx context.Context, input *dto.GetRevisionInput) (*dto.RevisionOutput, error) {
		return getRevision(ctx, application, input.GetInput, input.RevisionID)
	}, huma.OperationTags("conversation workflows"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-workflows", func(ctx context.Context, input *dto.ListInput) (*dto.ListOutput, error) {
		access, err := readAccess(ctx, application, input.WorkflowReadAccess)
		if err != nil {
			return nil, err
		}
		result, err := application.ListConversationWorkflows(ctx, app.ListWorkflowsInput{EvaluationRunAccess: access, Limit: input.Limit, Cursor: input.Cursor})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		rows := make([]dto.WorkflowHead, 0, len(result.Items))
		for _, head := range result.Items {
			rows = append(rows, mapper.HeadToResponse(head))
		}
		return &dto.ListOutput{Body: shared.SuccessEnvelope[[]dto.WorkflowHead]{Data: rows, Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.NextCursor != nil)}}, nil
	}, huma.OperationTags("conversation workflows"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions", func(ctx context.Context, input *dto.HistoryInput) (*dto.HistoryOutput, error) {
		access, err := readAccess(ctx, application, input.WorkflowReadAccess)
		if err != nil {
			return nil, err
		}
		result, err := application.ListConversationWorkflowRevisions(ctx, app.ListWorkflowRevisionsInput{EvaluationRunAccess: access, WorkflowID: model.WorkflowID(input.WorkflowID), Limit: input.Limit, Cursor: input.Cursor})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		rows := make([]dto.Revision, 0, len(result.Items))
		for _, revision := range result.Items {
			rows = append(rows, mapper.RevisionToResponse(revision))
		}
		return &dto.HistoryOutput{Body: shared.SuccessEnvelope[[]dto.Revision]{Data: rows, Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.NextCursor != nil)}}, nil
	}, huma.OperationTags("conversation workflows"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-workflow-selection", func(ctx context.Context, input *dto.WorkflowReadAccess) (*dto.SelectionOutput, error) {
		access, err := readAccess(ctx, application, *input)
		if err != nil {
			return nil, err
		}
		result, err := application.GetConversationWorkflowSelection(ctx, access)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.SelectionOutput{Body: shared.NullableSuccessEnvelope[dto.WorkflowSelection]{Data: mapper.SelectionToResponse(result), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("conversation workflows"), shared.SecuredOperation)
}
func readAccess(ctx context.Context, application app.App, input dto.WorkflowReadAccess) (app.EvaluationRunAccess, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return app.EvaluationRunAccess{}, err
	}
	return app.EvaluationRunAccess{Principal: principal, TenantID: tenant.ID(input.TenantID), Source: audit.SourceAPI, RequestID: input.RequestID}, nil
}
func getRevision(ctx context.Context, application app.App, input dto.GetInput, revisionID string) (*dto.RevisionOutput, error) {
	access, err := readAccess(ctx, application, input.WorkflowReadAccess)
	if err != nil {
		return nil, err
	}
	result, err := application.GetConversationWorkflow(ctx, app.GetWorkflowInput{EvaluationRunAccess: access, WorkflowID: model.WorkflowID(input.WorkflowID), RevisionID: model.WorkflowRevisionID(revisionID)})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return revisionOutput(result, input.TenantID), nil
}
