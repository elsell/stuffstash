package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationruns/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationruns/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/conversation-evaluation-runs", func(ctx context.Context, input *dto.QueueInput) (*dto.RunOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		cases := make([]app.EvaluationRunCaseReference, 0, len(input.Body.Cases))
		for _, reference := range input.Body.Cases {
			cases = append(cases, app.EvaluationRunCaseReference{CaseID: model.EvaluationCaseID(reference.CaseID), RevisionID: model.EvaluationCaseRevisionID(reference.RevisionID)})
		}
		run, err := application.QueueEvaluationRun(ctx, app.QueueEvaluationRunInput{EvaluationRunAccess: access, WorkflowID: model.WorkflowID(input.Body.WorkflowID), RevisionID: model.WorkflowRevisionID(input.Body.RevisionID), Cases: cases})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return runOutput(run, input.TenantID), nil
	}, huma.OperationTags("conversation evaluation runs"), shared.CreatedOperation, shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-evaluation-runs/{runId}", func(ctx context.Context, input *dto.GetInput) (*dto.RunOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		run, err := application.GetEvaluationRun(ctx, app.GetEvaluationRunInput{EvaluationRunAccess: access, RunID: model.EvaluationRunID(input.RunID)})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return runOutput(run, input.TenantID), nil
	}, huma.OperationTags("conversation evaluation runs"), shared.SecuredOperation)
	huma.Post(api, "/tenants/{tenantId}/conversation-evaluation-runs/{runId}/cancellation", func(ctx context.Context, input *dto.CancelInput) (*dto.RunOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		run, err := application.CancelEvaluationRun(ctx, app.CancelEvaluationRunInput{EvaluationRunAccess: access, RunID: model.EvaluationRunID(input.RunID), ExpectedVersion: input.Body.ExpectedVersion})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return runOutput(run, input.TenantID), nil
	}, huma.OperationTags("conversation evaluation runs"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/conversation-evaluation-runs", func(ctx context.Context, input *dto.ListInput) (*dto.ListOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		result, err := application.ListEvaluationRuns(ctx, app.ListEvaluationRunsInput{EvaluationRunAccess: access, Limit: input.Limit, Cursor: input.Cursor})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		rows := make([]dto.EvaluationRunHead, 0, len(result.Items))
		for _, head := range result.Items {
			rows = append(rows, mapper.HeadToResponse(head))
		}
		return &dto.ListOutput{Body: shared.SuccessEnvelope[[]dto.EvaluationRunHead]{Data: rows, Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.NextCursor != nil)}}, nil
	}, huma.OperationTags("conversation evaluation runs"), shared.SecuredOperation)
}
func authenticate(ctx context.Context, application app.App, input dto.AccessInput) (app.EvaluationRunAccess, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return app.EvaluationRunAccess{}, err
	}
	return app.EvaluationRunAccess{Principal: principal, TenantID: tenant.ID(input.TenantID), Source: audit.SourceAPI, RequestID: input.RequestID}, nil
}
func runOutput(run model.EvaluationRun, tenantID string) *dto.RunOutput {
	return &dto.RunOutput{Body: shared.SuccessEnvelope[dto.EvaluationRun]{Data: mapper.RunToResponse(run), Meta: shared.Meta{TenantID: tenantID}}}
}
