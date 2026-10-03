package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

func (r ConnectorRegistrar) Rotate(ctx context.Context, server, connectorID string) error {
	if strings.TrimSpace(connectorID) == "" {
		return ports.Failure("usage", "rotation requires --connector ID")
	}
	prior, err := r.Credentials.Load(ctx, server, connectorID)
	if err != nil {
		return err
	}
	if prior.Server != server || prior.ConnectorID != connectorID || prior.TenantID == "" || prior.InventoryID == "" {
		return ports.Failure("configuration", "stored connector identity does not match rotation target")
	}
	return r.pair(ctx, server, connectorID, nil, &prior)
}
