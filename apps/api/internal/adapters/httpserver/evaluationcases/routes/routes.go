package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/conversation-evaluation-cases", func(ctx context.Context, input *dto.CreateInput) (*dto.RevisionOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		revision, err := application.SaveEvaluationCaseRevision(ctx, app.SaveEvaluationCaseInput{EvaluationCaseAccess: access, Definition: mapper.DefinitionToDomain(input.Body.Definition)})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return revisionOutput(revision, input.TenantID), nil
	}, huma.OperationTags("conversation evaluation cases"), shared.CreatedOperation, shared.SecuredOperation)
	huma.Post(api, "/tenants/{tenantId}/conversation-evaluation-cases/{caseId}/revisions", func(ctx context.Context, input *dto.AppendInput) (*dto.RevisionOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		revision, err := application.SaveEvaluationCaseRevision(ctx, app.SaveEvaluationCaseInput{EvaluationCaseAccess: access, CaseID: domain.EvaluationCaseID(input.CaseID), ExpectedRevision: input.Body.ExpectedRevision, Definition: mapper.DefinitionToDomain(input.Body.Definition)})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return revisionOutput(revision, input.TenantID), nil
	}, huma.OperationTags("conversation evaluation cases"), shared.CreatedOperation, shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-evaluation-cases/{caseId}", func(ctx context.Context, input *dto.GetInput) (*dto.RevisionOutput, error) {
		return get(ctx, application, *input, "")
	}, huma.OperationTags("conversation evaluation cases"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-evaluation-cases/{caseId}/revisions/{revisionId}", func(ctx context.Context, input *dto.GetRevisionInput) (*dto.RevisionOutput, error) {
		return get(ctx, application, input.GetInput, input.RevisionID)
	}, huma.OperationTags("conversation evaluation cases"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-evaluation-cases", func(ctx context.Context, input *dto.ListInput) (*dto.ListOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		result, err := application.ListEvaluationCases(ctx, app.ListEvaluationCasesInput{EvaluationCaseAccess: access, Limit: input.Limit, Cursor: input.Cursor})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		rows := make([]dto.EvaluationCaseHead, 0, len(result.Items))
		for _, head := range result.Items {
			rows = append(rows, mapper.HeadToResponse(head))
		}
		return &dto.ListOutput{Body: shared.SuccessEnvelope[[]dto.EvaluationCaseHead]{Data: rows, Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.NextCursor != nil)}}, nil
	}, huma.OperationTags("conversation evaluation cases"), shared.SecuredOperation)
}
func authenticate(ctx context.Context, application app.App, input dto.AccessInput) (app.EvaluationCaseAccess, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return app.EvaluationCaseAccess{}, err
	}
	return app.EvaluationCaseAccess{Principal: principal, TenantID: tenant.ID(input.TenantID), Source: audit.SourceAPI, RequestID: input.RequestID}, nil
}
func get(ctx context.Context, application app.App, input dto.GetInput, revisionID string) (*dto.RevisionOutput, error) {
	access, err := authenticate(ctx, application, input.AccessInput)
	if err != nil {
		return nil, err
	}
	revision, err := application.GetEvaluationCase(ctx, app.GetEvaluationCaseInput{EvaluationCaseAccess: access, CaseID: domain.EvaluationCaseID(input.CaseID), RevisionID: domain.EvaluationCaseRevisionID(revisionID)})
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return revisionOutput(revision, input.TenantID), nil
}
func revisionOutput(value domain.EvaluationCaseRevision, tenantID string) *dto.RevisionOutput {
	return &dto.RevisionOutput{Body: shared.SuccessEnvelope[dto.EvaluationCaseRevision]{Data: mapper.RevisionToResponse(value), Meta: shared.Meta{TenantID: tenantID}}}
}
