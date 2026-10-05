package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"unicode/utf8"
)

func isPreferenceCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "notification-preferences"
}
func isPreferenceWrite(o Options) bool {
	return isPreferenceCommand(o) && len(o.Command) > 1 && o.Command[1] == "initialize"
}
func validatePreferences(o Options, scope bool) error {
	if len(o.Command) != 2 || (o.Command[1] != "show" && o.Command[1] != "initialize") {
		return ports.Failure("usage", "Use notification-preferences show or initialize --timezone ZONE.")
	}
	if o.Timezone != "" && !isPreferenceWrite(o) {
		return ports.Failure("usage", "Use --timezone with notification-preferences initialize.")
	}
	if utf8.RuneCountInString(o.Timezone) > 100 {
		return ports.Failure("usage", "The timezone name is too long. Use a timezone such as America/New_York.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.Page.Cursor != "" {
		return ports.Failure("usage", "Preference commands do not accept asset fields, cursors, or retry keys. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	return nil
}
func (r Runner) preparePreferenceInput(ctx context.Context, o Options) (Options, error) {
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
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory)); err != nil {
		return err
	}
	result, err := api.ChangeNotificationPreferences(ctx, o.Scope, ports.InitializePreferences, "", 0, o.RequestBody)
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.notification.preferences.initialized")
	return r.Output.Result(result)
}
