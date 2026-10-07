package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

type printResolutionInput struct {
	Revision    int64  `json:"revision"`
	Outcome     string `json:"reportedOutcome"`
	Acknowledge bool   `json:"acknowledgeUncertainty"`
}

func isPrintResolution(o Options) bool {
	return len(o.Command) > 1 && o.Command[0] == "print-jobs" && o.Command[1] == "resolve"
}
func decodePrintResolution(body []byte, requireRevision bool) (printResolutionInput, error) {
	var v printResolutionInput
	if json.Unmarshal(body, &v) != nil || (v.Outcome != "printed" && v.Outcome != "not_printed" && v.Outcome != "unknown") || !v.Acknowledge || (requireRevision && v.Revision <= 0) {
		return v, ports.Failure("usage", "Supply reportedOutcome (printed, not_printed or unknown), a positive revision, and acknowledgeUncertainty true in the input JSON.")
	}
	return v, nil
}
func (r Runner) preparePrintResolution(ctx context.Context, o Options) (Options, error) {
	if r.Picker == nil || o.JSON || o.NoInput {
		return o, ports.Failure("usage", "Supply --input FILE with the outcome, revision and uncertainty acknowledgment. Use --input - for JSON stdin.")
	}
	outcome, err := r.Picker.Pick(ctx, "What happened at the printer?", []ports.Choice{{ID: "unknown", Label: "Unknown", Detail: "Physical output is uncertain"}, {ID: "printed", Label: "Printed", Detail: "You observed the printed label"}, {ID: "not_printed", Label: "Not printed", Detail: "You observed that no label printed"}})
	if err != nil {
		return o, err
	}
	o.RequestBody, err = json.Marshal(printResolutionInput{Outcome: outcome, Acknowledge: true})
	if err != nil {
		return o, err
	}
	_, err = decodePrintResolution(o.RequestBody, false)
	return o, err
}
func (r Runner) resolvePrint(ctx context.Context, o Options, api ports.HumanPrintingAPI) (any, error) {
	v, err := decodePrintResolution(o.RequestBody, o.InputPath != "")
	if err != nil {
		return nil, err
	}
	if o.InputPath == "" {
		job, err := api.PrintJob(ctx, o.Scope, o.Command[2])
		if err != nil {
			return nil, err
		}
		if job.Data.Revision == 0 || job.Data.Revision > uint64(1<<63-1) {
			return nil, ports.Failure("protocol", "The job revision is not correct. Read the job again before you submit a resolution.")
		}
		v.Revision = int64(job.Data.Revision)
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Job: " + strconv.Quote(o.Command[2]) + ". Reported outcome: " + strconv.Quote(v.Outcome) + ". Revision: " + strconv.FormatInt(v.Revision, 10)); err != nil {
		return nil, err
	}
	if err := r.Output.Notice("This records your report. It does not verify physical output or print another label."); err != nil {
		return nil, err
	}
	if err := r.confirmAction(ctx, o, "Record print resolution", "Record", "Acknowledge uncertain physical output and save this report."); err != nil {
		return nil, err
	}
	body := o.RequestBody
	if o.InputPath == "" {
		body, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return api.ResolvePrint(ctx, o.Scope, o.Command[2], body)
}
