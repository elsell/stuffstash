package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// workflowEnvelope keeps nullable payloads explicit without nullable slice decoding.
type workflowEnvelope[T any] struct {
	Schema *string        `json:"$schema,omitempty"`
	Data   T              `json:"data"`
	Meta   generated.Meta `json:"meta"`
}

func workflowRevision(v generated.Revision) ports.WorkflowRevision {
	return ports.WorkflowRevision{AuthorId: v.AuthorId, CreatedAt: v.CreatedAt, Id: v.Id, Number: v.Number, SettingsMigration: v.SettingsMigration, WorkflowId: v.WorkflowId, Definition: ports.WorkflowDefinition{Budget: ports.WorkflowBudget(v.Definition.Budget), Instructions: v.Definition.Instructions, Name: v.Definition.Name, ProviderProfileId: v.Definition.ProviderProfileId}}
}
func (c *Client) Workflows(ctx context.Context, tenant string, p ports.Page) (ports.Result[[]ports.WorkflowHead], error) {
	r, err := read[workflowEnvelope[[]generated.WorkflowHead]](c.sdk.GetTenantsByTenantIdConversationWorkflows(ctx, tenant, &generated.GetTenantsByTenantIdConversationWorkflowsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.WorkflowHead]{}, err
	}
	var data []ports.WorkflowHead
	if r.Data != nil {
		data = make([]ports.WorkflowHead, len(r.Data))
	}
	for i, v := range r.Data {
		data[i] = ports.WorkflowHead(v)
	}
	return ports.Result[[]ports.WorkflowHead]{Data: data, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) Workflow(ctx context.Context, tenant, id string) (ports.Result[ports.WorkflowRevision], error) {
	r, err := read[generated.SuccessEnvelopeRevision](c.sdk.GetTenantsByTenantIdConversationWorkflowsByWorkflowId(ctx, tenant, id, nil))
	if err != nil {
		return ports.Result[ports.WorkflowRevision]{}, err
	}
	return ports.Result[ports.WorkflowRevision]{Data: workflowRevision(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) WorkflowRevisions(ctx context.Context, tenant, id string, p ports.Page) (ports.Result[[]ports.WorkflowRevision], error) {
	r, err := read[workflowEnvelope[[]generated.Revision]](c.sdk.GetTenantsByTenantIdConversationWorkflowsByWorkflowIdRevisions(ctx, tenant, id, &generated.GetTenantsByTenantIdConversationWorkflowsByWorkflowIdRevisionsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.WorkflowRevision]{}, err
	}
	var data []ports.WorkflowRevision
	if r.Data != nil {
		data = make([]ports.WorkflowRevision, len(r.Data))
	}
	for i, v := range r.Data {
		data[i] = workflowRevision(v)
	}
	return ports.Result[[]ports.WorkflowRevision]{Data: data, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) WorkflowRevision(ctx context.Context, tenant, id, revision string) (ports.Result[ports.WorkflowRevision], error) {
	r, err := read[generated.SuccessEnvelopeRevision](c.sdk.GetTenantsByTenantIdConversationWorkflowsByWorkflowIdRevisionsByRevisionId(ctx, tenant, id, revision, nil))
	if err != nil {
		return ports.Result[ports.WorkflowRevision]{}, err
	}
	return ports.Result[ports.WorkflowRevision]{Data: workflowRevision(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) WorkflowSelection(ctx context.Context, tenant string) (ports.Result[*ports.WorkflowSelection], error) {
	r, err := read[workflowEnvelope[*generated.WorkflowSelection]](c.sdk.GetTenantsByTenantIdConversationWorkflowSelection(ctx, tenant, nil))
	if err != nil {
		return ports.Result[*ports.WorkflowSelection]{}, err
	}
	var data *ports.WorkflowSelection
	if r.Data != nil {
		data = &ports.WorkflowSelection{WorkflowId: r.Data.WorkflowId, RevisionId: r.Data.RevisionId}
	}
	return ports.Result[*ports.WorkflowSelection]{Data: data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
