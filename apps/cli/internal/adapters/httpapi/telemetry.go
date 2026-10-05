package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) SubmitTelemetry(ctx context.Context, body json.RawMessage) (ports.Result[ports.TelemetryAccepted], error) {
	type envelope struct {
		Schema *string                 `json:"$schema,omitempty"`
		Data   ports.TelemetryAccepted `json:"data"`
		Meta   generated.Meta          `json:"meta"`
	}
	r, err := read[envelope](c.sdk.PostClientTelemetryWithBody(ctx, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.TelemetryAccepted]{}, err
	}
	return ports.Result[ports.TelemetryAccepted]{Schema: r.Schema, Data: r.Data, Meta: metadata(r.Meta)}, nil
}
