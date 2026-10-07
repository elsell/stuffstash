package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isDirectoryLifecycle(o Options) bool {
	return len(o.Command) == 2 && (o.Command[0] == "tenants" || o.Command[0] == "inventories") && (o.Command[1] == "archive" || o.Command[1] == "restore" || o.Command[1] == "delete")
}
func (r Runner) directoryLifecycle(ctx context.Context, o Options, session ports.Session) error {
	if o.IdempotencyKey != "" {
		return ports.Failure("usage", "This API operation does not support --idempotency-key. Remove the option.")
	}
	resource := ports.HouseholdResource
	target := "Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant)
	if o.Command[0] == "inventories" {
		resource = ports.InventoryResource
		target += ". Inventory: " + strconv.Quote(o.Scope.Inventory)
	}
	if err := r.Output.Notice(target); err != nil {
		return err
	}
	action := ports.LifecycleAction(o.Command[1])
	if action != ports.Restore && !o.Yes {

		detail := "Remove from active use. You can restore it later."
		if action == ports.Delete {
			detail = "Permanently delete this " + string(resource) + "."
		}
		if err := r.confirmAction(ctx, o, "Confirm "+string(action)+" "+string(resource), string(action), detail); err != nil {
			return err
		}
	}
	if r.DirectoryLifecycle == nil {
		return ports.Failure("configuration", "Lifecycle commands are not available. Update the CLI and try again.")
	}
	api, err := r.DirectoryLifecycle(o.Server, session.IDToken)
	if err != nil {
		return err
	}
	result, err := api.ChangeDirectoryLifecycle(ctx, resource, action, o.Scope)
	if err != nil {
		return err
	}
	if action == ports.Delete && r.Contexts != nil {
		inventory := o.Scope.Inventory
		if resource == ports.HouseholdResource {
			inventory = ""
		}
		if err := (contexts.Manager{Store: r.Contexts}).ClearResource(ctx, o.Server, contexts.Principal(session), o.Scope.Tenant, inventory); err != nil {
			_ = r.Output.Notice("The resource was deleted, but its saved scope could not be cleared. Use context list and context delete NAME to remove the stale selection.")
			r.Observer.Event(ctx, "cli.context.cleanup.failed")
		}
	}
	r.Observer.Event(ctx, "cli.directory.lifecycle.completed")
	return r.Output.Result(result)
}
