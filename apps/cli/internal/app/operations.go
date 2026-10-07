package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isOperationCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "operations" }
func validateOperations(o Options, scope bool) error {
	if len(o.Command) != 3 || o.Command[2] == "" || (o.Command[1] != "undo" && o.Command[1] != "redo") {
		return ports.Failure("usage", "Use operations undo OPERATION_ID or operations redo OPERATION_ID.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.Page.Cursor != "" {
		return ports.Failure("usage", "Undo and redo do not accept asset fields, cursors, or retry keys. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved context.")
	}
	return nil
}
func (r Runner) operationCommand(ctx context.Context, o Options, token string) error {
	if r.OperationsAPI == nil {
		return ports.Failure("configuration", "Operation commands are not available. Update the CLI and try again.")
	}
	api, err := r.OperationsAPI(o.Server, token)
	if err != nil {
		return err
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Operation: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	label, detail := "Undo", "Reverse this recorded operation."
	if o.Command[1] == "redo" {
		label, detail = "Redo", "Apply this recorded operation again."
	}
	if err := r.confirmAction(ctx, o, label+" operation", label, detail); err != nil {
		return err
	}
	result, err := api.ApplyOperation(ctx, o.Scope, o.Command[2], ports.OperationAction(o.Command[1]))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The operation result is unknown. Examine the affected asset with assets show ASSET_ID before you try again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.operation.applied")
	return r.Output.Result(result)
}
