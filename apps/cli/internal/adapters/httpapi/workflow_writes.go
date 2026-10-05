package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func workflowRevisionResult(response *http.Response, err error) (ports.Result[ports.WorkflowRevision], error) {
	r, err := read[generated.SuccessEnvelopeRevision](response, err)
	if err != nil {
		return ports.Result[ports.WorkflowRevision]{}, err
	}
	return ports.Result[ports.WorkflowRevision]{Data: workflowRevision(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) CreateWorkflow(ctx context.Context, tenant string, body []byte) (ports.Result[ports.WorkflowRevision], error) {
	return workflowRevisionResult(c.sdk.PostTenantsByTenantIdConversationWorkflowsWithBody(ctx, tenant, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) CreateWorkflowRevision(ctx context.Context, tenant, id string, body []byte) (ports.Result[ports.WorkflowRevision], error) {
	return workflowRevisionResult(c.sdk.PostTenantsByTenantIdConversationWorkflowsByWorkflowIdRevisionsWithBody(ctx, tenant, id, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) ActivateWorkflow(ctx context.Context, tenant, id string, body []byte) (ports.Result[ports.WorkflowRevision], error) {
	return workflowRevisionResult(c.sdk.PostTenantsByTenantIdConversationWorkflowsByWorkflowIdActivationWithBody(ctx, tenant, id, nil, "application/json", bytes.NewReader(body)))
}
