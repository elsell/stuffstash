package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) Principal(ctx context.Context) (ports.Result[ports.Principal], error) {
	r, err := read[generated.SuccessEnvelopePrincipalResponse](c.sdk.GetMe(ctx, nil))
	if err != nil {
		return ports.Result[ports.Principal]{}, err
	}
	return ports.Result[ports.Principal]{Data: ports.Principal{ID: r.Data.Id, DisplayName: r.Data.DisplayName, Email: r.Data.Email}}, nil
}
func (c *Client) Tenant(ctx context.Context, s ports.Scope) (ports.Result[ports.Tenant], error) {
	r, err := read[generated.SuccessEnvelopeTenantResponse](c.sdk.GetTenantsByTenantId(ctx, s.Tenant, nil))
	if err != nil {
		return ports.Result[ports.Tenant]{}, err
	}
	return ports.Result[ports.Tenant]{Data: ports.Tenant{ID: r.Data.Id, Name: r.Data.Name, Lifecycle: r.Data.LifecycleState, Access: access(r.Data.Access)}}, nil
}
func (c *Client) Inventory(ctx context.Context, s ports.Scope) (ports.Result[ports.Inventory], error) {
	r, err := read[generated.SuccessEnvelopeInventoryResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryId(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[ports.Inventory]{}, err
	}
	return ports.Result[ports.Inventory]{Data: inventory(r.Data)}, nil
}
func access(v generated.AccessResponse) ports.Access {
	return ports.Access{Relationship: v.Relationship, Permissions: v.Permissions.GetOrEmpty()}
}
func inventory(v generated.InventoryResponse) ports.Inventory {
	return ports.Inventory{ID: v.Id, TenantID: v.TenantId, Name: v.Name, Lifecycle: v.LifecycleState, Access: access(v.Access)}
}
