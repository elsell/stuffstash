package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type staticRealtimeVoiceProviderResolver struct {
	providers ports.RealtimeVoiceProviderSet
}

func (r staticRealtimeVoiceProviderResolver) ResolveRealtimeVoiceProviders(context.Context, ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	return r.providers, nil
}
