package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) ReviewPrintPairing(ctx context.Context, id string, body []byte) (ports.Result[ports.PairingReview], error) {
	return read[ports.Result[ports.PairingReview]](c.sdk.PostPrintConnectorPairingsByPairingIdReviewWithBody(ctx, id, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) ApprovePrintPairing(ctx context.Context, id string, body []byte) (ports.Result[ports.ApprovedConnector], error) {
	return read[ports.Result[ports.ApprovedConnector]](c.sdk.PostPrintConnectorPairingsByPairingIdApprovalWithBody(ctx, id, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) ApproveConnectorRotation(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.ApprovedPairing], error) {
	return read[ports.Result[ports.ApprovedPairing]](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintConnectorsByConnectorIdCredentialRotationWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
}
