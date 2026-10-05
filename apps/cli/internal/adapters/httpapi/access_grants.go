package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func accessGrant(v generated.GrantResponse) ports.AccessGrant {
	return ports.AccessGrant{TenantID: v.TenantId, InventoryID: v.InventoryId, PrincipalID: v.PrincipalId, Relationship: ports.AccessRelationship(v.Relationship)}
}
func (c *Client) AccessGrants(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.AccessGrant], error) {
	r, err := read[generated.SuccessEnvelopeListGrantResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAccessGrants(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdAccessGrantsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.AccessGrant]{}, err
	}
	var grants []ports.AccessGrant
	if r.Data.GetOrEmpty() != nil {
		grants = make([]ports.AccessGrant, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		grants = append(grants, accessGrant(v))
	}
	return ports.Result[[]ports.AccessGrant]{Data: grants, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) AccessGrant(ctx context.Context, s ports.Scope, principal string, role ports.AccessRelationship) (ports.Result[ports.AccessGrant], error) {
	r, err := read[generated.SuccessEnvelopeGrantResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAccessGrantsByPrincipalIdByRelationship(ctx, s.Tenant, s.Inventory, principal, generated.GetTenantsByTenantIdInventoriesByInventoryIdAccessGrantsByPrincipalIdByRelationshipParamsRelationship(role), nil))
	if err != nil {
		return ports.Result[ports.AccessGrant]{}, err
	}
	return ports.Result[ports.AccessGrant]{Data: accessGrant(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) RemoveAccessGrant(ctx context.Context, s ports.Scope, principal string, role ports.AccessRelationship) error {
	return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdAccessGrantsByPrincipalIdByRelationship(ctx, s.Tenant, s.Inventory, principal, generated.DeleteTenantsByTenantIdInventoriesByInventoryIdAccessGrantsByPrincipalIdByRelationshipParamsRelationship(role), nil))
}
