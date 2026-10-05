package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) Tenants(ctx context.Context, p ports.Page) (ports.Result[[]ports.Tenant], error) {
	response, err := read[generated.SuccessEnvelopeListMyTenantResponse](c.sdk.GetMeTenants(ctx, &generated.GetMeTenantsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Tenant]{}, err
	}
	items := make([]ports.Tenant, 0, len(response.Data.GetOrEmpty()))
	for _, item := range response.Data.GetOrEmpty() {
		items = append(items, ports.Tenant{ID: item.Id, Name: item.Name, Lifecycle: item.LifecycleState, Access: access(item.Access)})
	}
	return ports.Result[[]ports.Tenant]{Data: items, Pagination: page(response.Meta)}, nil
}
