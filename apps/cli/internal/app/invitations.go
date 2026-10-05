package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isInvitationCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "invitations" }
func validateInvitations(o Options, scope bool) error {
	valid := len(o.Command) == 2 && o.Command[1] == "list"
	if len(o.Command) == 3 && o.Command[2] != "" && (o.Command[1] == "show" || o.Command[1] == "cancel" || o.Command[1] == "delete") {
		valid = true
	}
	if !valid {
		return ports.Failure("usage", "Use invitations list, show ID, cancel ID, or delete ID.")
	}
	if o.InvitationStatus != "" {
		switch o.InvitationStatus {
		case "pending", "accepted", "cancelled", "expired", "revoked", "all":
		default:
			return ports.Failure("usage", "Choose pending, accepted, cancelled, expired, revoked, or all for --status.")
		}
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || (o.Command[1] != "list" && (o.Page.Cursor != "" || o.InvitationStatus != "")) {
		return ports.Failure("usage", "Invitation commands do not accept asset fields or retry keys. Use --status and --cursor only with invitations list.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	return nil
}
func (r Runner) invitationCommand(ctx context.Context, o Options, token string) error {
	if r.InvitationsAPI == nil {
		return ports.Failure("configuration", "Invitation commands are not available. Update the CLI and try again.")
	}
	api, err := r.InvitationsAPI(o.Server, token)
	if err != nil {
		return err
	}
	var result any
	switch o.Command[1] {
	case "list":
		result, err = api.Invitations(ctx, o.Scope, o.Page, o.InvitationStatus)
	case "show":
		result, err = api.Invitation(ctx, o.Scope, o.Command[2])
	case "cancel", "delete":
		if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; invitation: " + strconv.Quote(o.Command[2])); err != nil {
			return err
		}
		label := "Cancel invitation"
		status := "cancelled"
		if o.Command[1] == "delete" {
			label = "Delete invitation"
			status = "deleted"
		}
		if err := r.confirmAction(ctx, o, label, label, "Change this invitation only. Existing access grants remain separate."); err != nil {
			return err
		}
		err = api.ChangeInvitation(ctx, o.Scope, o.Command[2], ports.InvitationAction(o.Command[1]))
		result = map[string]string{"status": status, "tenantId": o.Scope.Tenant, "inventoryId": o.Scope.Inventory, "invitationId": o.Command[2]}
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.invitation."+o.Command[1]+".completed")
	return r.Output.Result(result)
}
