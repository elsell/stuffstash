package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

func (c *Client) CreateInvitation(ctx context.Context, s ports.Scope, body []byte) (ports.Result[ports.CreatedInvitation], error) {
	r, err := read[generated.SuccessEnvelopeCreatedInvitationResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.CreatedInvitation]{}, err
	}
	v := r.Data
	if strings.TrimSpace(v.Id) == "" || strings.TrimSpace(v.InviteUrl) == "" {
		return ports.Result[ports.CreatedInvitation]{}, ports.Failure("protocol", "The server did not return the invitation ID and link.")
	}
	return ports.Result[ports.CreatedInvitation]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.CreatedInvitation{Invitation: ports.Invitation{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, Email: v.Email, Relationship: ports.AccessRelationship(v.Relationship), Status: string(v.Status), ExpiresAt: v.ExpiresAt, IsExpired: v.IsExpired, InviterPrincipalID: v.InviterPrincipalId, AcceptedPrincipalID: v.AcceptedPrincipalId}, InviteURL: v.InviteUrl}}, nil
}
