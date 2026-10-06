package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) CreateEvaluationCase(ctx context.Context, tenant string, body []byte) (ports.Result[ports.EvaluationCaseRevision], error) {
	return readEvaluation[ports.EvaluationCaseRevision](c.sdk.PostTenantsByTenantIdConversationEvaluationCasesWithBody(ctx, tenant, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) CreateEvaluationRevision(ctx context.Context, tenant, id string, body []byte) (ports.Result[ports.EvaluationCaseRevision], error) {
	return readEvaluation[ports.EvaluationCaseRevision](c.sdk.PostTenantsByTenantIdConversationEvaluationCasesByCaseIdRevisionsWithBody(ctx, tenant, id, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) CreateEvaluationRun(ctx context.Context, tenant string, body []byte) (ports.Result[ports.EvaluationRun], error) {
	return readEvaluation[ports.EvaluationRun](c.sdk.PostTenantsByTenantIdConversationEvaluationRunsWithBody(ctx, tenant, nil, "application/json", bytes.NewReader(body)))
}
