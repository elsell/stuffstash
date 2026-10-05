package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

// Decode declared evaluation fields directly: generated nullable decoding would
// collapse optional strings and float64 would round high precision durations.
func readEvaluation[T any](response *http.Response, err error) (ports.Result[T], error) {
	result, err := read[ports.Result[T]](response, err)
	if result.Meta != nil {
		result.Pagination = result.Meta.Pagination
	}
	return result, err
}

func (c *Client) EvaluationCases(ctx context.Context, tenant string, p ports.Page) (ports.Result[[]ports.EvaluationCaseHead], error) {
	return readEvaluation[[]ports.EvaluationCaseHead](c.sdk.GetTenantsByTenantIdConversationEvaluationCases(ctx, tenant, &generated.GetTenantsByTenantIdConversationEvaluationCasesParams{Limit: &p.Limit, Cursor: &p.Cursor}))
}

func (c *Client) EvaluationCase(ctx context.Context, tenant, id string) (ports.Result[ports.EvaluationCaseRevision], error) {
	return readEvaluation[ports.EvaluationCaseRevision](c.sdk.GetTenantsByTenantIdConversationEvaluationCasesByCaseId(ctx, tenant, id, nil))
}

func (c *Client) EvaluationRevisions(ctx context.Context, tenant, id string, p ports.Page) (ports.Result[[]ports.EvaluationCaseRevision], error) {
	return readEvaluation[[]ports.EvaluationCaseRevision](c.sdk.GetTenantsByTenantIdConversationEvaluationCasesByCaseIdRevisions(ctx, tenant, id, &generated.GetTenantsByTenantIdConversationEvaluationCasesByCaseIdRevisionsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
}

func (c *Client) EvaluationRevision(ctx context.Context, tenant, id, revision string) (ports.Result[ports.EvaluationCaseRevision], error) {
	return readEvaluation[ports.EvaluationCaseRevision](c.sdk.GetTenantsByTenantIdConversationEvaluationCasesByCaseIdRevisionsByRevisionId(ctx, tenant, id, revision, nil))
}

func (c *Client) EvaluationRuns(ctx context.Context, tenant string, p ports.Page) (ports.Result[[]ports.EvaluationRunHead], error) {
	return readEvaluation[[]ports.EvaluationRunHead](c.sdk.GetTenantsByTenantIdConversationEvaluationRuns(ctx, tenant, &generated.GetTenantsByTenantIdConversationEvaluationRunsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
}

func (c *Client) EvaluationRun(ctx context.Context, tenant, id string) (ports.Result[ports.EvaluationRun], error) {
	return readEvaluation[ports.EvaluationRun](c.sdk.GetTenantsByTenantIdConversationEvaluationRunsByRunId(ctx, tenant, id, nil))
}
