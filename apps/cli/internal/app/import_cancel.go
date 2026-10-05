package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isImportCancel(o Options) bool {
	return isImportJobCommand(o) && len(o.Command) > 1 && o.Command[1] == "cancel"
}
func importCancellationMode(body []byte) (ports.ImportCancellationMode, error) {
	var v struct {
		Mode ports.ImportCancellationMode `json:"mode"`
	}
	if json.Unmarshal(body, &v) != nil || (v.Mode != ports.KeepImportProgress && v.Mode != ports.DiscardImportProgress) {
		return "", ports.Failure("usage", "Set mode to keep_partial_progress or discard_partial_progress in the input JSON.")
	}
	return v.Mode, nil
}
func (r Runner) prepareImportCancel(ctx context.Context, o Options) (Options, error) {
	if r.Picker == nil || o.NoInput || o.JSON {
		return o, ports.Failure("usage", "Supply --input FILE with mode set to keep_partial_progress or discard_partial_progress.")
	}
	mode, err := r.Picker.Pick(ctx, "Cancel import", []ports.Choice{{ID: string(ports.KeepImportProgress), Label: "Keep partial progress", Detail: "Stop the job and keep imported records"}, {ID: string(ports.DiscardImportProgress), Label: "Discard partial progress", Detail: "Stop the job and remove partial imported records"}})
	if err != nil {
		return o, err
	}
	if mode != string(ports.KeepImportProgress) && mode != string(ports.DiscardImportProgress) {
		return o, context.Canceled
	}
	o.RequestBody, err = json.Marshal(map[string]string{"mode": mode})
	return o, err
}
func (r Runner) cancelImport(ctx context.Context, o Options, api ports.ImportJobsAPI) error {
	mode, err := importCancellationMode(o.RequestBody)
	if err != nil {
		return err
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; import job: " + strconv.Quote(o.Command[2]) + "; mode: " + strconv.Quote(string(mode))); err != nil {
		return err
	}
	detail := "Stop this import and keep imported records."
	if mode == ports.DiscardImportProgress {
		detail = "Stop this import and remove partial imported records."
	}
	if err := r.confirmAction(ctx, o, "Cancel import", "Cancel import", detail); err != nil {
		return err
	}
	result, err := api.CancelImportJob(ctx, o.Scope, o.Command[2], o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The cancellation result is unknown. Run import-jobs show JOB_ID before you retry.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.import_job.cancellation.requested")
	return r.Output.Result(result)
}
