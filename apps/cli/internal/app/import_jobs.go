package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isImportJobCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "import-jobs" }
func validateImportJobs(o Options, scope bool) error {
	if !(len(o.Command) == 2 && o.Command[1] == "preview") && !(len(o.Command) == 3 && o.Command[1] == "start" && o.Command[2] != "") && !(len(o.Command) == 2 && o.Command[1] == "list") && !(len(o.Command) == 3 && o.Command[2] != "" && (o.Command[1] == "cancel" || o.Command[1] == "show" || o.Command[1] == "delete")) {
		return ports.Failure("usage", "Use import-jobs list, show JOB_ID, cancel JOB_ID, delete JOB_ID, preview, or start JOB_ID.")
	}
	if o.IdempotencyKey != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" || o.Page.Cursor != "" {
		return ports.Failure("usage", "Import job history does not accept asset fields, retry keys or cursors. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	return nil
}
func (r Runner) importJobCommand(ctx context.Context, o Options, token string) error {
	if r.ImportJobsAPI == nil {
		return ports.Failure("configuration", "Import job commands are not available. Update the CLI and try again.")
	}
	api, err := r.ImportJobsAPI(o.Server, token)
	if err != nil {
		return err
	}
	var result any
	switch o.Command[1] {
	case "cancel":
		return r.cancelImport(ctx, o, api)
	case "list":
		result, err = api.ImportJobs(ctx, o.Scope)
	case "show":
		result, err = api.ImportJob(ctx, o.Scope, o.Command[2])
	case "delete":
		if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; import job: " + strconv.Quote(o.Command[2])); err != nil {
			return err
		}
		if err := r.confirmAction(ctx, o, "Remove import job from history", "Remove", "Remove this job from history. Keep imported assets."); err != nil {
			return err
		}
		err = api.DeleteImportJob(ctx, o.Scope, o.Command[2])
		result = map[string]string{"status": "removed", "tenantId": o.Scope.Tenant, "inventoryId": o.Scope.Inventory, "jobId": o.Command[2]}
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.import_job."+o.Command[1]+".completed")
	return r.Output.Result(result)
}
