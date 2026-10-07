package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isDirectoryCommand(o Options) bool {
	return len(o.Command) == 2 && ((o.Command[0] == "tenants" && (o.Command[1] == "list" || o.Command[1] == "show")) || ((o.Command[0] == "account" || o.Command[0] == "inventories") && o.Command[1] == "show"))
}
func isAccountCommand(o Options) bool {
	return isTenantList(o) || len(o.Command) == 2 && o.Command[0] == "account" && o.Command[1] == "show"
}
func requiresInventory(o Options) bool {
	if isArchiveCommand(o) {
		return len(o.Command) > 1 && o.Command[1] == "create"
	}
	if isSearch(o) && o.AllInventories {
		return false
	}
	if isCustomization(o) {
		return o.DefinitionLevel == "inventory"
	}
	if isProviderProfileCommand(o) || isEvaluationCommand(o) || isWorkflowCommand(o) || isVoiceProviderCommand(o) {
		return false
	}
	return !(len(o.Command) == 2 && (o.Command[0] == "tenants" || o.Command[0] == "inventories" && (o.Command[1] == "list" || o.Command[1] == "create")))
}
func (r Runner) directoryCommand(ctx context.Context, o Options, token string) error {
	if r.DirectoryAPI == nil {
		return ports.Failure("configuration", "Account and household commands are not available. Update the CLI and try again.")
	}
	api, err := r.DirectoryAPI(o.Server, token)
	if err != nil {
		return err
	}
	var result any
	switch {
	case isTenantList(o):
		result, err = listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.Tenant], error) { return api.Tenants(ctx, page) })
	case o.Command[0] == "account":
		result, err = api.Principal(ctx)
	case o.Command[0] == "tenants":
		result, err = api.Tenant(ctx, o.Scope)
	default:
		result, err = api.Inventory(ctx, o.Scope)
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.directory.read.completed")
	return r.Output.Result(result)
}
