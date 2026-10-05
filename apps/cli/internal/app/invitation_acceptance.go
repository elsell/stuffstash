package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

func isInvitationTokenCommand(o Options) bool {
	return isInvitationCommand(o) && len(o.Command) > 1 && (o.Command[1] == "preview" || o.Command[1] == "accept")
}
func validateInvitationToken(body []byte) error {
	var v struct {
		Token string `json:"acceptanceToken"`
	}
	if json.Unmarshal(body, &v) != nil || strings.TrimSpace(v.Token) == "" {
		return ports.Failure("usage", "Supply a nonempty acceptanceToken in the input JSON.")
	}
	return nil
}
func (r Runner) prepareInvitationToken(ctx context.Context, o Options) (Options, error) {
	if r.SecretInput == nil || o.NoInput || o.JSON {
		return o, ports.Failure("usage", "Supply --input FILE with acceptanceToken. Use --input - to read JSON from stdin.")
	}
	token, err := r.SecretInput.ReadSecret(ctx, "Invitation acceptance token", 4095)
	if err != nil {
		return o, err
	}
	o.RequestBody, err = json.Marshal(map[string]string{"acceptanceToken": token})
	if err != nil {
		return o, err
	}
	return o, validateInvitationToken(o.RequestBody)
}
func (r Runner) invitationTokenCommand(ctx context.Context, o Options, api ports.InvitationsAPI) error {
	preview, err := api.PreviewInvitation(ctx, o.Scope, o.Command[2], o.RequestBody)
	if err != nil {
		return err
	}
	if o.Command[1] == "preview" {
		r.Observer.Event(ctx, "cli.invitation.previewed")
		return r.Output.Result(preview)
	}
	v := preview.Data
	if v.InventoryID != o.Scope.Inventory || (v.Relationship != ports.AccessViewer && v.Relationship != ports.AccessEditor) {
		return ports.Failure("protocol", "The invitation preview is not valid. Check the invitation details and try again.")
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(v.InventoryName) + " (" + strconv.Quote(v.InventoryID) + "); invitation: " + strconv.Quote(o.Command[2]) + "; role: " + strconv.Quote(string(v.Relationship)) + "; status: " + strconv.Quote(v.Status) + "; expires: " + strconv.Quote(v.ExpiresAt) + "; expired: " + strconv.FormatBool(v.IsExpired)); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Accept invitation", "Accept", "Join the displayed inventory with the displayed role."); err != nil {
		return err
	}
	result, err := api.AcceptInvitation(ctx, o.Scope, o.Command[2], o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The acceptance result is unknown. Run invitations preview INVITATION_ID with the same token before you retry.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.invitation.accepted")
	return r.Output.Result(result)
}
