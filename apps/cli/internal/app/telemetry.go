package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"slices"
	"strconv"
)

func isTelemetry(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "telemetry" }
func validateTelemetry(o Options) error {
	if len(o.Command) != 2 || o.Command[1] != "submit" {
		return ports.Failure("usage", "Use telemetry submit --input FILE to send existing client measurements.")
	}
	return nil
}

type telemetryInput struct {
	Schema       *string `json:"$schema,omitempty"`
	Measurements []struct {
		Platform  string   `json:"platform"`
		Operation string   `json:"operation"`
		Surface   string   `json:"surface"`
		Variant   string   `json:"variant"`
		Outcome   string   `json:"outcome"`
		Duration  *float64 `json:"durationMs"`
	} `json:"measurements"`
}

func decodeTelemetry(o Options) (telemetryInput, error) {
	var input telemetryInput
	if o.InputPath == "" {
		return input, ports.Failure("usage", "Supply --input FILE or --input - with existing client measurements.")
	}
	d := json.NewDecoder(bytes.NewReader(o.RequestBody))
	d.DisallowUnknownFields()
	if !json.Valid(o.RequestBody) || d.Decode(&input) != nil || len(input.Measurements) < 1 || len(input.Measurements) > 50 {
		return input, ports.Failure("usage", "Supply a JSON object with one to 50 measurements. Use telemetry submit --help for the fields.")
	}
	for _, m := range input.Measurements {
		if !slices.Contains([]string{"ios", "android", "web"}, m.Platform) || !slices.Contains([]string{"request", "image"}, m.Operation) || !slices.Contains([]string{"application", "home", "list", "detail", "gallery", "fullscreen", "upload"}, m.Surface) || !slices.Contains([]string{"none", "small", "medium", "large", "original"}, m.Variant) || !slices.Contains([]string{"success", "failure", "cancelled"}, m.Outcome) || m.Duration == nil || *m.Duration < 0 || *m.Duration > 60000 {
			return input, ports.Failure("usage", "A measurement has a missing or incorrect field. Use telemetry submit --help for supported values and duration limits.")
		}
	}
	return input, nil
}
func prepareTelemetry(o Options) (Options, error) { _, err := decodeTelemetry(o); return o, err }
func (r Runner) telemetryCommand(ctx context.Context, o Options, token string) error {
	input, err := decodeTelemetry(o)
	if err != nil {
		return err
	}
	if r.TelemetryAPI == nil {
		return ports.Failure("configuration", "Telemetry submission is not available. Update the CLI and try again.")
	}
	if err = r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Measurements: " + strconv.Itoa(len(input.Measurements)) + ". Send only measurements that you intend to record."); err != nil {
		return err
	}
	if err = r.confirmAction(ctx, o, "Submit client measurements", "Submit measurements", "A repeated submission can count the same measurements again."); err != nil {
		return err
	}
	api, err := r.TelemetryAPI(o.Server, token)
	if err != nil {
		return err
	}
	result, err := api.SubmitTelemetry(ctx, o.RequestBody)
	if err != nil {
		var failure *ports.Error
		if errors.As(err, &failure) && (failure.Category == "network" || failure.Category == "unavailable" || failure.Category == "api" || failure.Category == "protocol") {
			return ports.Failure(failure.Category, "The submission result is unknown. Examine server telemetry before you send the batch again. The server can count it twice.")
		}
		return err
	}
	r.Observer.Event(ctx, "cli.telemetry.submitted")
	return r.Output.Result(result)
}
