package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func registerHistory(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/conversation-evaluation-cases/{caseId}/revisions", func(ctx context.Context, input *dto.HistoryInput) (*dto.HistoryOutput, error) {
		access, err := authenticate(ctx, application, input.AccessInput)
		if err != nil {
			return nil, err
		}
		result, err := application.ListEvaluationCaseRevisions(ctx, app.ListEvaluationCaseRevisionsInput{EvaluationCaseAccess: access, CaseID: domain.EvaluationCaseID(input.CaseID), Limit: input.Limit, Cursor: input.Cursor})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		rows := make([]dto.EvaluationCaseRevision, 0, len(result.Items))
		for _, value := range result.Items {
			rows = append(rows, mapper.RevisionToResponse(value))
		}
		return &dto.HistoryOutput{Body: shared.SuccessEnvelope[[]dto.EvaluationCaseRevision]{Data: rows, Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.NextCursor != nil)}}, nil
	}, huma.OperationTags("conversation evaluation cases"), shared.SecuredOperation)
}
