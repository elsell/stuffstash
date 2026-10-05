package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

type grantInput struct {
	PrincipalID  string                   `json:"principalId"`
	Relationship ports.AccessRelationship `json:"relationship"`
}

func isGrantCreate(o Options) bool {
	return isAccessGrantCommand(o) && len(o.Command) > 1 && o.Command[1] == "create"
}
func decodeGrant(body []byte) (grantInput, error) {
	var v grantInput
	if json.Unmarshal(body, &v) != nil || strings.TrimSpace(v.PrincipalID) == "" || (v.Relationship != ports.AccessViewer && v.Relationship != ports.AccessEditor) {
		return v, ports.Failure("usage", "Supply principalId and relationship (viewer or editor) in the grant input.")
	}
	return v, nil
}
func (r Runner) prepareGrantInput(ctx context.Context, o Options) (Options, error) {
	if r.TextInput == nil || r.Picker == nil || o.NoInput || o.JSON {
		return o, ports.Failure("usage", "Supply --input FILE with principalId and relationship (viewer or editor).")
	}
	principal, err := r.TextInput.ReadText(ctx, "Principal ID", 1024)
	if err != nil {
		return o, err
	}
	role, err := r.Picker.Pick(ctx, "Inventory access", []ports.Choice{{ID: "viewer", Label: "Viewer", Detail: "Read inventory items"}, {ID: "editor", Label: "Editor", Detail: "Read and change inventory items"}})
	if err != nil {
		return o, err
	}
	if role != "viewer" && role != "editor" {
		return o, context.Canceled
	}
	o.RequestBody, err = json.Marshal(grantInput{principal, ports.AccessRelationship(role)})
	if err != nil {
		return o, err
	}
	_, err = decodeGrant(o.RequestBody)
	return o, err
}
func (r Runner) createGrant(ctx context.Context, o Options, api ports.AccessGrantsAPI) error {
	v, err := decodeGrant(o.RequestBody)
	if err != nil {
		return err
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; principal: " + strconv.Quote(v.PrincipalID) + "; relationship: " + strconv.Quote(string(v.Relationship))); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Grant inventory access", "Grant access", "Add this relationship to the selected inventory."); err != nil {
		return err
	}
	result, err := api.CreateAccessGrant(ctx, o.Scope, o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The grant result is unknown. Run access-grants show PRINCIPAL_ID viewer|editor for this relationship before you retry.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.access_grant.created")
	return r.Output.Result(result)
}
