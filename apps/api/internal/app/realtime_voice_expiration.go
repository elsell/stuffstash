package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
)

type realtimeVoiceExpiration = agentapp.RealtimeVoiceExpiration

func (a App) realtimeVoiceExpiration(ctx context.Context, session RealtimeVoiceSession, item asset.Asset) (*realtimeVoiceExpiration, error) {
	return a.realtimeReadTools().RealtimeVoiceExpiration(ctx, realtimeReadScope(session), item)
}
