package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isTagCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "tags" }
func isTagWrite(o Options) bool {
	return isTagCommand(o) && len(o.Command) > 1 && (o.Command[1] == "create" || o.Command[1] == "update")
}
func acceptsBody(o Options) bool {
	return isImportCancel(o) || isInvitationTokenCommand(o) || isInvitationExpiration(o) || isGrantCreate(o) || isDeviceRegistration(o) || isPreferenceWrite(o) || isDirectoryWrite(o) || isTagWrite(o) || isAssetWrite(o) || isCheckoutWrite(o)
}
func validateTags(o Options, requireScope bool) error {
	if len(o.Command) < 2 {
		return ports.Failure("usage", "Supply a tag action. Use tags list, create, update ID, or delete ID.")
	}
	action := o.Command[1]
	valid := (len(o.Command) == 2 && (action == "list" || action == "create")) || (len(o.Command) == 3 && (action == "update" || action == "delete") && o.Command[2] != "")
	if !valid {
		return ports.Failure("usage", "The tag command is invalid. Use --help to check the arguments.")
	}
	if (action == "list" || action == "delete") && (o.ConnectorName != "" || o.TagColor != nil) {
		return ports.Failure("usage", "Use --name and --tag-color only with tags create or tags update.")
	}
	if o.TagKey != nil && action != "create" {
		return ports.Failure("usage", "Use --key only with tags create. A saved tag key cannot change.")
	}
	if requireScope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	if action != "list" && o.IdempotencyKey != "" {
		return ports.Failure("usage", "This API operation does not support --idempotency-key. Remove the option.")
	}
	return nil
}
func (r Runner) prepareTagInput(ctx context.Context, o Options) (Options, error) {
	fields := map[string]string{}
	name := o.ConnectorName
	if name == "" && (o.Command[1] == "create" || o.TagColor == nil) {
		if r.TextInput == nil || o.JSON || o.NoInput {
			return o, ports.Failure("usage", "Supply --name NAME or --input FILE for this command.")
		}
		var err error
		name, err = r.TextInput.ReadText(ctx, "Tag name", 80)
		if err != nil {
			return o, err
		}
	}
	if name != "" {
		fields["displayName"] = name
	}
	if o.TagKey != nil {
		fields["key"] = *o.TagKey
	}
	if o.TagColor != nil {
		fields["color"] = *o.TagColor
	}
	o.RequestBody, _ = json.Marshal(fields)
	return o, nil
}
func (r Runner) tagsCommand(ctx context.Context, o Options, token string) error {
	if r.TagsAPI == nil {
		return ports.Failure("configuration", "Tag commands are not available. Update the CLI and try again.")
	}
	api, err := r.TagsAPI(o.Server, token)
	if err != nil {
		return err
	}
	if o.Command[1] == "list" {
		result, err := api.Tags(ctx, o.Scope, o.Page)
		if err != nil {
			return err
		}
		return r.Output.Result(result)
	}
	id := ""
	if len(o.Command) == 3 {
		id = o.Command[2]
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; tag: " + strconv.Quote(id)); err != nil {
		return err
	}
	if o.Command[1] == "delete" {
		if err := r.confirmAction(ctx, o, "Confirm tag deletion", "Delete", "Delete this tag from the inventory."); err != nil {
			return err
		}
	}
	result, err := api.ChangeTag(ctx, o.Scope, ports.TagAction(o.Command[1]), id, o.RequestBody)
	if err != nil {
		var failure *ports.Error
		if o.Command[1] == "create" && errors.As(err, &failure) && (failure.Category == "network" || failure.Category == "protocol" || failure.Category == "unavailable" || failure.Category == "api") {
			return ports.Failure(failure.Category, "The create result is unknown. Run tags list before you retry.")
		}
		return err
	}
	r.Observer.Event(ctx, "cli.tag.command.completed")
	return r.Output.Result(result)
}
