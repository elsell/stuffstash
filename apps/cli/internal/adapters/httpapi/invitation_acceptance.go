package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PreviewInvitation(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.InvitationPreview], error) {
	r, err := read[generated.SuccessEnvelopeInvitationPreviewResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsByInvitationIdPreviewWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.InvitationPreview]{}, err
	}
	v := r.Data
	return ports.Result[ports.InvitationPreview]{Data: ports.InvitationPreview{InventoryID: v.InventoryId, InventoryName: v.InventoryName, Relationship: ports.AccessRelationship(v.Relationship), Status: string(v.Status), ExpiresAt: v.ExpiresAt, IsExpired: v.IsExpired}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) AcceptInvitation(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.InvitationAcceptance], error) {
	r, err := read[generated.SuccessEnvelopeInvitationAcceptanceResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAccessInvitationsByInvitationIdAcceptWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.InvitationAcceptance]{}, err
	}
	return ports.Result[ports.InvitationAcceptance]{Data: ports.InvitationAcceptance{Invitation: invitation(r.Data.Invitation), Grant: accessGrant(r.Data.Grant)}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
