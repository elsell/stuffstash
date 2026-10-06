package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) ResolvePrint(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdResolutionWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body))))
}
