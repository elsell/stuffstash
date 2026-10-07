package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isEvaluationCancellation(o Options) bool {
	return len(o.Command) >= 3 && o.Command[0] == "evaluation" && o.Command[1] == "runs" && o.Command[2] == "cancel"
}
func decodeEvaluationCancellation(body []byte) (ports.EvaluationCancellation, error) {
	var input ports.EvaluationCancellation
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if !json.Valid(body) || decoder.Decode(&input) != nil || input.ExpectedVersion <= 0 {
		return input, ports.Failure("usage", "Supply a positive integer expectedVersion in the input JSON. Only expectedVersion and optional $schema are accepted.")
	}
	return input, nil
}
func (r Runner) prepareEvaluationCancellation(o Options) (Options, error) {
	interactive := r.Picker != nil && !o.NoInput && !o.JSON
	if o.InputPath == "" {
		if !interactive || o.Yes {
			return o, ports.Failure("usage", "Supply --input FILE with expectedVersion and add --yes. Use --input - for JSON stdin.")
		}
		return o, nil
	}
	if _, err := decodeEvaluationCancellation(o.RequestBody); err != nil {
		return o, err
	}
	if !interactive && !o.Yes {
		return o, ports.Failure("usage", "Examine the run and expectedVersion. Add --yes to approve cancellation.")
	}
	return o, nil
}
func (r Runner) cancelEvaluationRun(ctx context.Context, o Options, api ports.EvaluationAPI) error {
	var input ports.EvaluationCancellation
	if o.InputPath != "" {
		var err error
		input, err = decodeEvaluationCancellation(o.RequestBody)
		if err != nil {
			return err
		}
	} else {
		result, err := api.EvaluationRun(ctx, o.Scope.Tenant, o.Command[3])
		if err != nil {
			return err
		}
		if result.Data.Version <= 0 {
			return ports.Failure("protocol", "The server returned an incorrect run version. Run evaluation runs show RUN_ID before you cancel.")
		}
		input.ExpectedVersion = result.Data.Version
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Run: " + strconv.Quote(o.Command[3]) + ". Version: " + strconv.FormatInt(input.ExpectedVersion, 10)); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Cancel evaluation run", "Cancel run", "Request cancellation of this evaluation run."); err != nil {
		return err
	}
	result, err := api.CancelEvaluationRun(ctx, o.Scope.Tenant, o.Command[3], input)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "conflict":
				return ports.Failure("conflict", "The server rejected the cancellation. Run evaluation runs show RUN_ID. Examine its version and state before you try again.")
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The cancellation result is unknown. Run evaluation runs show RUN_ID before you try again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.evaluation_run.cancellation.requested")
	return r.Output.Result(result)
}
