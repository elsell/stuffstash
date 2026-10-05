package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) WriteDirectory(ctx context.Context, kind ports.DirectoryWrite, s ports.Scope, body []byte) (any, error) {
	switch kind {
	case ports.CreateTenant:
		r, err := read[generated.SuccessEnvelopeTenantResponse](c.sdk.PostTenantsWithBody(ctx, nil, "application/json", bytes.NewReader(body)))
		if err != nil {
			return nil, err
		}
		return ports.Result[ports.Tenant]{Schema: r.Schema, Meta: metadata(r.Meta), Data: tenant(r.Data)}, nil
	case ports.UpdateTenant:
		r, err := read[generated.SuccessEnvelopeTenantResponse](c.sdk.PatchTenantsByTenantIdWithBody(ctx, s.Tenant, nil, "application/json", bytes.NewReader(body)))
		if err != nil {
			return nil, err
		}
		return ports.Result[ports.Tenant]{Schema: r.Schema, Meta: metadata(r.Meta), Data: tenant(r.Data)}, nil
	case ports.CreateInventory:
		r, err := read[generated.SuccessEnvelopeInventoryResponse](c.sdk.PostTenantsByTenantIdInventoriesWithBody(ctx, s.Tenant, nil, "application/json", bytes.NewReader(body)))
		if err != nil {
			return nil, err
		}
		return ports.Result[ports.Inventory]{Schema: r.Schema, Meta: metadata(r.Meta), Data: inventory(r.Data)}, nil
	case ports.UpdateInventory:
		r, err := read[generated.SuccessEnvelopeInventoryResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
		if err != nil {
			return nil, err
		}
		return ports.Result[ports.Inventory]{Schema: r.Schema, Meta: metadata(r.Meta), Data: inventory(r.Data)}, nil
	default:
		return nil, ports.Failure("usage", "Unknown directory action. Use --help to choose a command.")
	}
}
func tenant(v generated.TenantResponse) ports.Tenant {
	return ports.Tenant{ID: v.Id, Name: v.Name, Lifecycle: v.LifecycleState, Access: access(v.Access)}
}
