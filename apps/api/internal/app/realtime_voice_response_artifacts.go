package app

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
)

const maxRealtimeVoiceResponseArtifacts = 16

func validateRealtimeVoiceResponseArtifacts(displayResponse string, artifacts []ports.StructuredAgentResponseArtifact) error {
	if len(artifacts) > maxRealtimeVoiceResponseArtifacts {
		return ports.ErrInvalidProviderInput
	}
	seen := map[asset.ID]struct{}{}
	for _, artifact := range artifacts {
		id, validID := asset.NewID(artifact.AssetID.String())
		title, validTitle := asset.NewTitle(artifact.Title)
		kind, validKind := asset.NewKind(artifact.AssetKind.String())
		context := strings.TrimSpace(artifact.Context)
		_, validContext := asset.NewTitle(context)
		if artifact.Context == "" {
			validContext = true
		}
		if artifact.Type != ports.StructuredAgentResponseArtifactAssetReference || !validID || id != artifact.AssetID ||
			!validTitle || title.String() != strings.TrimSpace(artifact.Title) || !validKind || kind != artifact.AssetKind ||
			!validContext || context != artifact.Context {
			return ports.ErrInvalidProviderInput
		}
		if _, duplicate := seen[artifact.AssetID]; duplicate {
			return ports.ErrInvalidProviderInput
		}
		seen[artifact.AssetID] = struct{}{}
	}
	return nil
}
