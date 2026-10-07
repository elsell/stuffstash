package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isAuditCommand(o Options) bool {
	return len(o.Command) > 1 && o.Command[1] == "audit" && (o.Command[0] == "tenants" || o.Command[0] == "inventories" || o.Command[0] == "assets")
}
func validateAudit(o Options, scope bool) error {
	if o.Command[0] == "assets" {
		if len(o.Command) != 3 || o.Command[2] == "" {
			return ports.Failure("usage", "Use assets audit ASSET_ID.")
		}
		if o.Page.Cursor != "" {
			return ports.Failure("usage", "Asset audit history does not support --cursor. Remove the option.")
		}
	} else if len(o.Command) != 2 {
		return ports.Failure("usage", "Use tenants audit or inventories audit with the selected context.")
	}
	if o.IdempotencyKey != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" {
		return ports.Failure("usage", "Audit reads do not accept asset fields or retry keys. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant for household history. For inventory or asset history, also supply --inventory. You can select a saved context instead.")
	}
	return nil
}
func (r Runner) auditCommand(ctx context.Context, o Options, token string) error {
	if r.AuditAPI == nil {
		return ports.Failure("configuration", "Audit commands are not available. Update the CLI and try again.")
	}
	api, err := r.AuditAPI(o.Server, token)
	if err != nil {
		return err
	}
	q := ports.AuditQuery{Scope: o.Scope, Page: o.Page, Level: ports.InventoryAudit}
	if o.Command[0] == "tenants" {
		q.Level = ports.HouseholdAudit
	}
	if o.Command[0] == "assets" {
		q.Level = ports.AssetAudit
		q.AssetID = o.Command[2]
	}
	result, err := api.AuditRecords(ctx, q)
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.audit.read.completed")
	return r.Output.Result(result)
}
