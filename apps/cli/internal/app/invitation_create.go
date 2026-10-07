package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/mail"
	"strconv"
	"strings"
)

type invitationCreateInput struct {
	Email        string `json:"email"`
	Relationship string `json:"relationship"`
}

func isInvitationCreate(o Options) bool {
	return len(o.Command) == 2 && o.Command[0] == "invitations" && o.Command[1] == "create"
}
func decodeInvitationCreate(body []byte) (invitationCreateInput, error) {
	var v invitationCreateInput
	if json.Unmarshal(body, &v) != nil {
		return v, ports.Failure("usage", "Supply email and relationship in a JSON object.")
	}
	parsed, err := mail.ParseAddress(v.Email)
	if err != nil || parsed.Address != v.Email || strings.TrimSpace(v.Email) == "" {
		return v, ports.Failure("usage", "Supply a correct email address without a display name.")
	}
	if v.Relationship != "viewer" && v.Relationship != "editor" {
		return v, ports.Failure("usage", "Select viewer or editor for the invitation role.")
	}
	return v, nil
}
func (r Runner) prepareInvitationCreate(ctx context.Context, o Options) (Options, error) {
	v := invitationCreateInput{Email: o.InvitationEmail, Relationship: o.InvitationRole}
	var err error
	if !o.NoInput && !o.JSON {
		if v.Email == "" && r.TextInput != nil {
			v.Email, err = r.TextInput.ReadText(ctx, "Invitee email address", 320)
			if err != nil {
				return o, err
			}
		}
		if v.Relationship == "" && r.Picker != nil {
			v.Relationship, err = r.Picker.Pick(ctx, "Inventory access", []ports.Choice{{ID: "viewer", Label: "Viewer", Detail: "Read inventory items"}, {ID: "editor", Label: "Editor", Detail: "Read and change inventory items"}})
			if err != nil {
				return o, err
			}
		}
	}
	if v.Email == "" || v.Relationship == "" {
		return o, ports.Failure("usage", "Supply --email ADDRESS and --role viewer|editor, or --input FILE|- with email and relationship.")
	}
	o.RequestBody, err = json.Marshal(v)
	if err != nil {
		return o, err
	}
	_, err = decodeInvitationCreate(o.RequestBody)
	return o, err
}
func (r Runner) createInvitation(ctx context.Context, o Options, token string) error {
	v, err := decodeInvitationCreate(o.RequestBody)
	if err != nil {
		return err
	}
	if r.InvitationWriter == nil {
		return ports.Failure("configuration", "Invitation creation is not available. Update the CLI and try again.")
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Email: " + strconv.Quote(v.Email) + ". Role: " + v.Relationship); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Create inventory invitation", "Create invitation", "Allow this email address to accept the selected inventory role."); err != nil {
		return err
	}
	api, err := r.InvitationWriter(o.Server, token)
	if err != nil {
		return err
	}
	result, err := api.CreateInvitation(ctx, o.Scope, o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) && (failure.Category == "network" || failure.Category == "protocol" || failure.Category == "unavailable") {
			return ports.Failure(failure.Category, "The invitation result is unknown. Run invitations list before you try again. A lost invitation link cannot be shown again.")
		}
		return err
	}
	r.Observer.Event(ctx, "cli.invitation.created")
	return r.Output.Result(result)
}
