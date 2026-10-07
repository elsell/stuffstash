package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"time"
)

func isInvitationExpiration(o Options) bool {
	return isInvitationCommand(o) && len(o.Command) > 1 && o.Command[1] == "expiration"
}
func invitationExpiration(body []byte) (string, error) {
	var v struct {
		ExpiresAt string `json:"expiresAt"`
	}
	if json.Unmarshal(body, &v) != nil {
		return "", ports.Failure("usage", "Supply expiresAt as an RFC3339 timestamp with a timezone.")
	}
	if _, err := time.Parse(time.RFC3339Nano, v.ExpiresAt); err != nil {
		return "", ports.Failure("usage", "Use an RFC3339 timestamp with a timezone, such as 2030-01-01T12:00:00Z.")
	}
	return v.ExpiresAt, nil
}
func (r Runner) prepareInvitationExpiration(ctx context.Context, o Options) (Options, error) {
	if r.TextInput == nil || o.NoInput || o.JSON {
		return o, ports.Failure("usage", "Supply --input FILE with expiresAt as an RFC3339 timestamp.")
	}
	value, err := r.TextInput.ReadText(ctx, "Expiration timestamp with timezone (RFC3339)", 64)
	if err != nil {
		return o, err
	}
	o.RequestBody, err = json.Marshal(map[string]string{"expiresAt": value})
	if err != nil {
		return o, err
	}
	_, err = invitationExpiration(o.RequestBody)
	return o, err
}
func (r Runner) updateInvitationExpiration(ctx context.Context, o Options, api ports.InvitationsAPI) error {
	expires, err := invitationExpiration(o.RequestBody)
	if err != nil {
		return err
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Invitation: " + strconv.Quote(o.Command[2]) + ". Expiration: " + strconv.Quote(expires)); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Change invitation expiration", "Change expiration", expires); err != nil {
		return err
	}
	result, err := api.UpdateInvitationExpiration(ctx, o.Scope, o.Command[2], o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The update result is unknown. Run invitations show INVITATION_ID before you try again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.invitation.expiration.updated")
	return r.Output.Result(result)
}
