package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r Runner) serverCommand(ctx context.Context, o Options) error {
	if len(o.Command) != 2 || (o.Command[1] != "show" && o.Command[1] != "auth-config") {
		return ports.Failure("usage", "Use server show or server auth-config.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.PrintLabel || o.Page.Cursor != "" || o.TagColor != nil || o.TagKey != nil || o.Details != nil {
		return ports.Failure("usage", "Server discovery does not accept mutation fields or cursors. Remove those options.")
	}
	if r.ServerAPI == nil {
		return ports.Failure("configuration", "Server discovery is not available. Update the CLI and try again.")
	}
	api, err := r.ServerAPI(o.Server)
	if err != nil {
		return err
	}
	var result any
	if o.Command[1] == "show" {
		result, err = api.ServerInfo(ctx)
	} else {
		result, err = api.ServerAuthConfig(ctx)
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.server.discovery.completed")
	return r.Output.Result(result)
}
