package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isProviderProfileCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "provider-profiles"
}
func validateProviderProfiles(o Options, scope bool) error {
	if isProviderWrite(o) {
		return validateProviderWrite(o, scope)
	}
	if !(len(o.Command) == 2 && o.Command[1] == "list") && !(len(o.Command) == 3 && (o.Command[1] == "show" || isProviderAction(o.Command[1])) && o.Command[2] != "") {
		return ports.Failure("usage", "Use provider-profiles list, or show|enable|disable|archive|test PROFILE_ID.")
	}
	if o.IdempotencyKey != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" || o.Page.Cursor != "" || o.Details != nil || o.TagColor != nil || o.TagKey != nil {
		return ports.Failure("usage", "These provider profile commands do not accept input fields or cursors. Remove those options.")
	}
	if scope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "Supply --tenant, or choose a saved household context.")
	}
	return nil
}
func (r Runner) providerProfileCommand(ctx context.Context, o Options, token string) error {
	if isProviderWrite(o) {
		return r.writeProvider(ctx, o, token)
	}
	if r.ProviderProfilesAPI == nil {
		return ports.Failure("configuration", "Provider profile commands are not available. Update the CLI and try again.")
	}
	api, err := r.ProviderProfilesAPI(o.Server, token)
	if err != nil {
		return err
	}
	if isProviderAction(o.Command[1]) {
		return r.providerAction(ctx, o, api)
	}
	var result any
	if o.Command[1] == "list" {
		result, err = api.ProviderProfiles(ctx, o.Scope.Tenant)
	} else {
		result, err = api.ProviderProfile(ctx, o.Scope.Tenant, o.Command[2])
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.provider_profile.read.completed")
	return r.Output.Result(result)
}
