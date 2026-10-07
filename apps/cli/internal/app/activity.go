package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isActivityCommand(o Options) bool {
	return len(o.Command) > 1 && o.Command[0] == "assets" && o.Command[1] == "activity"
}
func validateActivity(o Options, scope bool) error {
	if len(o.Command) != 3 || o.Command[2] == "" {
		return ports.Failure("usage", "Use assets activity ASSET_ID.")
	}
	if o.ActivityView != "" && o.ActivityView != ports.ActivityChanges && o.ActivityView != ports.ActivityAll {
		return ports.Failure("usage", "Select changes or all for --view.")
	}
	if o.IdempotencyKey != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" {
		return ports.Failure("usage", "Activity reads do not accept asset fields or retry keys. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved context.")
	}
	return nil
}
func (r Runner) activityCommand(ctx context.Context, o Options, token string) error {
	if r.ActivityAPI == nil {
		return ports.Failure("configuration", "Activity commands are not available. Update the CLI and try again.")
	}
	api, err := r.ActivityAPI(o.Server, token)
	if err != nil {
		return err
	}
	result, err := listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.Activity], error) {
		return api.AssetActivity(ctx, o.Scope, o.Command[2], o.ActivityView, page)
	})
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.asset.activity.read.completed")
	return r.Output.Result(result)
}
