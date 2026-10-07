package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isProviderAction(action string) bool {
	switch action {
	case "enable", "disable", "archive", "test":
		return true
	}
	return false
}
func (r Runner) providerAction(ctx context.Context, o Options, api ports.ProviderProfilesAPI) error {
	action := o.Command[1]
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Provider profile: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	detail := "Change the provider profile state."
	switch action {
	case "enable":
		detail = "Enable this provider profile."
	case "disable":
		detail = "Disable this provider profile. It will not be available for use."
	case "archive":
		detail = "Archive this provider profile. It will not be available for use."
	case "test":
		detail = "Contact the configured provider and record the test result."
	}
	if err := r.confirmAction(ctx, o, "Provider profile: "+action, action, detail); err != nil {
		return err
	}
	var result any
	var err error
	switch action {
	case "enable":
		result, err = api.EnableProvider(ctx, o.Scope.Tenant, o.Command[2])
	case "disable":
		result, err = api.DisableProvider(ctx, o.Scope.Tenant, o.Command[2])
	case "archive":
		result, err = api.ArchiveProvider(ctx, o.Scope.Tenant, o.Command[2])
	case "test":
		result, err = api.TestProvider(ctx, o.Scope.Tenant, o.Command[2])
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.provider_profile."+action+".completed")
	return r.Output.Result(result)
}
