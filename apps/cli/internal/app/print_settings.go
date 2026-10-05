package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isPrintSettingsCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "print-settings"
}
func validatePrintSettings(o Options, scope bool) error {
	if len(o.Command) != 2 || o.Command[1] != "show" {
		return ports.Failure("usage", "Use print-settings show.")
	}
	if o.IdempotencyKey != "" || o.Page.Cursor != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" {
		return ports.Failure("usage", "Print settings inspection does not accept mutation fields or cursors. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	return nil
}
func (r Runner) printSettingsCommand(ctx context.Context, o Options, token string) error {
	if r.PrintSettingsAPI == nil {
		return ports.Failure("configuration", "Print settings inspection is not available. Update the CLI and try again.")
	}
	api, err := r.PrintSettingsAPI(o.Server, token)
	if err != nil {
		return err
	}
	result, err := api.PrintSettings(ctx, o.Scope)
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.print_settings.read.completed")
	return r.Output.Result(result)
}
