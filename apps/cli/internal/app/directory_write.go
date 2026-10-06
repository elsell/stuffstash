package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isDirectoryWrite(o Options) bool {
	return len(o.Command) == 2 && (o.Command[0] == "tenants" || o.Command[0] == "inventories") && (o.Command[1] == "create" || o.Command[1] == "update")
}
func isTenantCreate(o Options) bool {
	return isDirectoryWrite(o) && o.Command[0] == "tenants" && o.Command[1] == "create"
}
func (r Runner) writeDirectory(ctx context.Context, o Options, token string) error {
	if r.DirectoryWriter == nil {
		return ports.Failure("configuration", "Household and inventory writes are not available. Update the CLI and try again.")
	}
	api, err := r.DirectoryWriter(o.Server, token)
	if err != nil {
		return err
	}
	target := "Server: " + strconv.Quote(o.Server)
	if !isTenantCreate(o) {
		target += "; household: " + strconv.Quote(o.Scope.Tenant)
	}
	if o.Command[0] == "inventories" && o.Command[1] == "update" {
		target += "; inventory: " + strconv.Quote(o.Scope.Inventory)
	}
	if err := r.Output.Notice(target); err != nil {
		return err
	}

	kind := ports.CreateTenant
	if o.Command[0] == "tenants" && o.Command[1] == "update" {
		kind = ports.UpdateTenant
	}
	if o.Command[0] == "inventories" {
		kind = ports.CreateInventory
		if o.Command[1] == "update" {
			kind = ports.UpdateInventory
		}
	}
	result, err := api.WriteDirectory(ctx, kind, o.Scope, o.RequestBody)
	if err != nil {
		var failure *ports.Error
		if o.Command[1] == "create" && errors.As(err, &failure) && (failure.Category == "network" || failure.Category == "protocol" || failure.Category == "unavailable" || failure.Category == "api") {
			return ports.Failure(failure.Category, "The create result is unknown. Run "+o.Command[0]+" list before you retry.")
		}
		return err
	}
	r.Observer.Event(ctx, "cli.directory.write.completed")
	return r.Output.Result(result)
}
