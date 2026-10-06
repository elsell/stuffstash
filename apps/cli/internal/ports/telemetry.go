package ports

import (
	"context"
	"encoding/json"
)

type TelemetryAccepted struct {
	Accepted int `json:"accepted"`
}
type TelemetryAPI interface {
	SubmitTelemetry(context.Context, json.RawMessage) (Result[TelemetryAccepted], error)
}
