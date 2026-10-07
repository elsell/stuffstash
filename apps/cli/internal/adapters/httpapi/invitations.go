package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func invitation(v generated.InvitationResponse) ports.Invitation {
	return ports.Invitation{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, Email: v.Email, Relationship: ports.AccessRelationship(v.Relationship), Status: string(v.Status), ExpiresAt: v.ExpiresAt, IsExpired: v.IsExpired, InviterPrincipalID: v.InviterPrincipalId, AcceptedPrincipalID: v.AcceptedPrincipalId}
}
func (c *Client) Invitations(ctx context.Context, s ports.Scope, p ports.Page, status string) (ports.Result[[]ports.Invitation], error) {
	params := &generated.GetTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsParams{Limit: &p.Limit, Cursor: &p.Cursor}
	if status != "" {
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsParamsStatus(status)
		params.Status = &v
	}
	r, err := read[generated.SuccessEnvelopeListInvitationResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAccessInvitations(ctx, s.Tenant, s.Inventory, params))
	if err != nil {
		return ports.Result[[]ports.Invitation]{}, err
	}
	var items []ports.Invitation
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.Invitation, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, invitation(v))
	}
	return ports.Result[[]ports.Invitation]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) Invitation(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.Invitation], error) {
	r, err := read[generated.SuccessEnvelopeInvitationResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsByInvitationId(ctx, s.Tenant, s.Inventory, id, nil))
	if err != nil {
		return ports.Result[ports.Invitation]{}, err
	}
	return ports.Result[ports.Invitation]{Data: invitation(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) ChangeInvitation(ctx context.Context, s ports.Scope, id string, a ports.InvitationAction) error {
	switch a {
	case ports.CancelInvitation:
		return noContent(c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsByInvitationIdCancel(ctx, s.Tenant, s.Inventory, id, nil))
	case ports.DeleteInvitation:
		return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsByInvitationId(ctx, s.Tenant, s.Inventory, id, nil))
	default:
		return ports.Failure("usage", "Select cancel or delete for the invitation.")
	}
}

func (c *Client) UpdateInvitationExpiration(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.Invitation], error) {
	r, err := read[generated.SuccessEnvelopeInvitationResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsByInvitationIdExpirationWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.Invitation]{}, err
	}
	return ports.Result[ports.Invitation]{Data: invitation(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
