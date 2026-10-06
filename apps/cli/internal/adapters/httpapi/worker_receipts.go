package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) recordReceipt(ctx context.Context, operation string, result ports.ProtocolReceiptResult) {
	if c.receipts != nil {
		c.receipts.Record(ctx, ports.ProtocolReceipt{Operation: operation, Result: result})
	}
}
func (c *Client) recordAttempt(ctx context.Context, operation string, result ports.Result[workerAttempt]) {
	var data *ports.ConsumerAttempt
	if !result.Data.null {
		data = &result.Data.ConsumerAttempt
	}
	c.recordReceipt(ctx, operation, ports.WorkerAttemptReceipt{Data: data, Schema: result.Schema, Meta: result.Meta, Pagination: result.Pagination})
}
