package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isConnectorInspection(o Options) bool {
	return len(o.Command) >= 3 && o.Command[0] == "connectors" && o.Command[1] == "print" && (o.Command[2] == "list" || o.Command[2] == "show")
}
func validateConnectorInspection(o Options, scope bool) error {
	if !(len(o.Command) == 3 && o.Command[2] == "list") && !(len(o.Command) == 4 && o.Command[2] == "show" && o.Command[3] != "") {
		return ports.Failure("usage", "Use connectors print list or connectors print show CONNECTOR_ID.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || (o.Command[2] == "show" && o.Page.Cursor != "") {
		return ports.Failure("usage", "Connector inspection does not accept mutation fields or a detail cursor. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	return nil
}
func (r Runner) connectorInspection(ctx context.Context, o Options, token string) error {
	if r.ConnectorInspectionAPI == nil {
		return ports.Failure("configuration", "Connector inspection is not available. Update the CLI and try again.")
	}
	api, err := r.ConnectorInspectionAPI(o.Server, token)
	if err != nil {
		return err
	}
	var result any
	if o.Command[2] == "list" {
		result, err = api.PrintConnectors(ctx, o.Scope, o.Page)
	} else {
		result, err = api.PrintConnector(ctx, o.Scope, o.Command[3])
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.print_connector.read.completed")
	return r.Output.Result(result)
}
