package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) ImportJobs(ctx context.Context, s ports.Scope) (ports.Result[ports.ImportJobList], error) {
	r, err := read[generated.SuccessEnvelopeImportJobListResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdImportsJobs(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[ports.ImportJobList]{}, err
	}
	var jobs []ports.ImportJob
	if r.Data.Jobs.GetOrEmpty() != nil {
		jobs = make([]ports.ImportJob, 0, len(r.Data.Jobs.GetOrEmpty()))
	}
	for _, v := range r.Data.Jobs.GetOrEmpty() {
		jobs = append(jobs, mapImportJob(v))
	}
	return ports.Result[ports.ImportJobList]{Data: ports.ImportJobList{Jobs: jobs}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) ImportJob(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.ImportJob], error) {
	r, err := read[generated.SuccessEnvelopeImportJobResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdImportsJobsByJobId(ctx, s.Tenant, s.Inventory, id, nil))
	if err != nil {
		return ports.Result[ports.ImportJob]{}, err
	}
	return ports.Result[ports.ImportJob]{Data: mapImportJob(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) DeleteImportJob(ctx context.Context, s ports.Scope, id string) error {
	return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdImportsJobsByJobId(ctx, s.Tenant, s.Inventory, id, nil))
}

func (c *Client) CancelImportJob(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.ImportJob], error) {
	r, err := read[generated.SuccessEnvelopeImportJobResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdImportsJobsByJobIdCancelWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.ImportJob]{}, err
	}
	return ports.Result[ports.ImportJob]{Data: mapImportJob(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
