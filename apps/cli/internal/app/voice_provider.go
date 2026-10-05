package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isVoiceProviderCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "voice-provider"
}
func validateVoiceProvider(o Options, scope bool) error {
	if len(o.Command) != 2 || (o.Command[1] != "show" && o.Command[1] != "update") {
		return ports.Failure("usage", "Use voice-provider show or update [--input FILE|-] for household voice selections.")
	}
	if scope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "Supply --tenant, or choose a saved household context.")
	}
	return nil
}
func (r Runner) voiceProviderCommand(ctx context.Context, o Options, token string) error {
	if r.VoiceProviderAPI == nil {
		return ports.Failure("configuration", "Voice provider inspection is not available. Update the CLI and try again.")
	}
	api, err := r.VoiceProviderAPI(o.Server, token)
	if err != nil {
		return err
	}
	if isVoiceProviderUpdate(o) {
		return r.updateVoiceProvider(ctx, o, token, api)
	}
	result, err := api.VoiceProviderConfiguration(ctx, o.Scope.Tenant)
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.voice_provider.read.completed")
	return r.Output.Result(result)
}
