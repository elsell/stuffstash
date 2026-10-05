package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

// The generated routes are retained, but their response types narrow uint32 and
// uint64 values. Decode declared port models to preserve the server contract.
func readConsumerInspection[T any](response *http.Response, err error) (ports.Result[T], error) {
	result, err := read[ports.Result[T]](response, err)
	if result.Meta != nil {
		result.Pagination = result.Meta.Pagination
	}
	return result, consumerError(err)
}
func (c *Client) ConsumerPrinters(ctx context.Context) (ports.Result[[]ports.ConsumerPrinter], error) {
	return readConsumerInspection[[]ports.ConsumerPrinter](c.sdk.GetPrintConsumerPrinters(ctx, nil))
}
func (c *Client) ConsumerAttempts(ctx context.Context, p ports.Page, printer, status string) (ports.Result[[]ports.ConsumerAttempt], error) {
	filter := generated.GetPrintConsumerAttemptsParamsStatus(status)
	return readConsumerInspection[[]ports.ConsumerAttempt](c.sdk.GetPrintConsumerAttempts(ctx, &generated.GetPrintConsumerAttemptsParams{PrinterId: &printer, Status: &filter, Limit: &p.Limit, Cursor: &p.Cursor}))
}
func (c *Client) ConsumerAttempt(ctx context.Context, id string) (ports.Result[*ports.ConsumerAttempt], error) {
	return readConsumerInspection[*ports.ConsumerAttempt](c.sdk.GetPrintConsumerAttemptsByAttemptId(ctx, id, nil))
}
