package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isAccessGrantCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "access-grants"
}
func validateAccessGrants(o Options, scope bool) error {
	valid := len(o.Command) == 2 && (o.Command[1] == "list" || o.Command[1] == "create")
	if len(o.Command) == 4 && (o.Command[1] == "show" || o.Command[1] == "remove") && o.Command[2] != "" && (o.Command[3] == string(ports.AccessViewer) || o.Command[3] == string(ports.AccessEditor)) {
		valid = true
	}
	if !valid {
		return ports.Failure("usage", "Use access-grants list, create, show PRINCIPAL_ID viewer|editor, or remove PRINCIPAL_ID viewer|editor.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || (o.Command[1] != "list" && o.Page.Cursor != "") {
		return ports.Failure("usage", "Access grants do not accept asset fields or retry keys. Use --cursor only with access-grants list.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	return nil
}
func (r Runner) accessGrantCommand(ctx context.Context, o Options, token string) error {
	if r.AccessGrantsAPI == nil {
		return ports.Failure("configuration", "Access grant commands are not available. Update the CLI and try again.")
	}
	api, err := r.AccessGrantsAPI(o.Server, token)
	if err != nil {
		return err
	}
	var result any
	switch o.Command[1] {
	case "create":
		return r.createGrant(ctx, o, api)
	case "list":
		result, err = api.AccessGrants(ctx, o.Scope, o.Page)
	case "show":
		result, err = api.AccessGrant(ctx, o.Scope, o.Command[2], ports.AccessRelationship(o.Command[3]))
	case "remove":
		target := "Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; principal: " + strconv.Quote(o.Command[2]) + "; relationship: " + strconv.Quote(o.Command[3])
		if err := r.Output.Notice(target); err != nil {
			return err
		}
		if err := r.confirmAction(ctx, o, "Remove access grant", "Remove", "Remove this relationship. Other access grants may still apply."); err != nil {
			return err
		}
		err = api.RemoveAccessGrant(ctx, o.Scope, o.Command[2], ports.AccessRelationship(o.Command[3]))
		result = map[string]string{"status": "removed", "tenantId": o.Scope.Tenant, "inventoryId": o.Scope.Inventory, "principalId": o.Command[2], "relationship": o.Command[3]}
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.access_grant."+o.Command[1]+".completed")
	return r.Output.Result(result)
}
