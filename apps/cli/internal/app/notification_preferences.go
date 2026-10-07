package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"unicode/utf8"
)

func isPreferenceCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "notification-preferences"
}
func isPreferenceWrite(o Options) bool {
	return isPreferenceCommand(o) && len(o.Command) > 1 && (o.Command[1] == "initialize" || o.Command[1] == "update" || o.Command[1] == "override")
}
func validatePreferences(o Options, scope bool) error {
	valid := len(o.Command) == 2 && (o.Command[1] == "show" || o.Command[1] == "initialize" || o.Command[1] == "update")
	if len(o.Command) == 3 && o.Command[2] != "" && (o.Command[1] == "override" || o.Command[1] == "remove-override") {
		valid = true
	}
	if !valid {
		return ports.Failure("usage", "Use notification-preferences show, initialize, update, override TYPE_ID, or remove-override TYPE_ID.")
	}
	if o.Timezone != "" && o.Command[1] != "initialize" {
		return ports.Failure("usage", "Use --timezone with notification-preferences initialize.")
	}
	if o.Revision != -1 && (o.Command[1] != "remove-override" || o.Revision < 1) {
		return ports.Failure("usage", "Use a positive --revision with remove-override. Put write revisions in the JSON input.")
	}
	if utf8.RuneCountInString(o.Timezone) > 100 {
		return ports.Failure("usage", "The timezone name is too long. Use a timezone such as America/New_York.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.Page.Cursor != "" {
		return ports.Failure("usage", "Preference commands do not accept asset fields, cursors, or retry keys. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved context.")
	}
	return nil
}
func (r Runner) preparePreferenceInput(ctx context.Context, o Options) (Options, error) {
	if o.Command[1] != "initialize" {
		if r.Picker == nil || r.TextInput == nil || o.JSON || o.NoInput {
			return o, ports.Failure("usage", "Supply --input FILE with the complete settings and revision for this command.")
		}
		return o, nil
	}
	zone := o.Timezone
	if zone == "" {
		if r.TextInput == nil || o.JSON || o.NoInput {
			return o, ports.Failure("usage", "Supply --timezone ZONE or --input FILE to initialize preferences.")
		}
		var err error
		zone, err = r.TextInput.ReadText(ctx, "Timezone (for example, America/New_York)", 100)
		if err != nil {
			return o, err
		}
	}
	if zone == "" || utf8.RuneCountInString(zone) > 100 {
		return o, ports.Failure("usage", "Supply a timezone name with 1 to 100 characters.")
	}
	o.RequestBody, _ = json.Marshal(map[string]string{"timezone": zone})
	return o, nil
}
func (r Runner) preferencesCommand(ctx context.Context, o Options, token string) error {
	if r.NotificationPreferencesAPI == nil {
		return ports.Failure("configuration", "Preference commands are not available. Update the CLI and try again.")
	}
	api, err := r.NotificationPreferencesAPI(o.Server, token)
	if err != nil {
		return err
	}
	if o.Command[1] == "show" {
		result, err := api.NotificationPreferences(ctx, o.Scope)
		if err != nil {
			return err
		}
		return r.Output.Result(result)
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory)); err != nil {
		return err
	}
	id := ""
	if len(o.Command) == 3 {
		id = o.Command[2]
		if err := r.Output.Notice("Asset type: " + strconv.Quote(id)); err != nil {
			return err
		}
	}
	revision := o.Revision
	if o.Command[1] == "remove-override" {
		if revision == -1 {
			if r.TextInput == nil || o.JSON || o.NoInput {
				return ports.Failure("usage", "Supply --revision N from notification-preferences show.")
			}
			text, err := r.TextInput.ReadText(ctx, "Preference revision", 19)
			if err != nil {
				return err
			}
			revision, err = strconv.ParseInt(text, 10, 64)
			if err != nil || revision < 1 {
				return ports.Failure("usage", "The preference revision must be a positive integer.")
			}
		}
		if err := r.confirmAction(ctx, o, "Remove notification override", "Remove", "Use the default policy for this asset type."); err != nil {
			return err
		}
	} else if len(o.RequestBody) == 0 {
		current, err := api.NotificationPreferences(ctx, o.Scope)
		if err != nil {
			return err
		}
		o.RequestBody, err = r.editPreferenceBody(ctx, o, current.Data)
		if err != nil {
			return err
		}
	}
	result, err := api.ChangeNotificationPreferences(ctx, o.Scope, ports.PreferenceAction(o.Command[1]), id, revision, o.RequestBody)
	if err != nil {
		var failure *ports.Error
		if errors.As(err, &failure) && failure.Category == "conflict" {
			return ports.Failure("conflict", "Notification preferences changed. Run notification-preferences show and review the current settings before you try again.")
		}
		return err
	}
	r.Observer.Event(ctx, "cli.notification.preferences.changed")
	return r.Output.Result(result)
}
