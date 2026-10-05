package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func importJobResult(response *http.Response, err error) (ports.Result[ports.ImportJob], error) {
	r, err := read[generated.SuccessEnvelopeImportJobResponse](response, err)
	if err != nil {
		return ports.Result[ports.ImportJob]{}, err
	}
	return ports.Result[ports.ImportJob]{Data: mapImportJob(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) PreviewImportJob(ctx context.Context, s ports.Scope, body []byte) (ports.Result[ports.ImportJob], error) {
	return importJobResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdImportsJobsPreviewWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) StartImportJob(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.ImportJob], error) {
	return importJobResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdImportsJobsByJobIdStartWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
}
