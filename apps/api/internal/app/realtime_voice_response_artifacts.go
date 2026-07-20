package app

import (
	"regexp"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxRealtimeVoiceResponseArtifacts = 16

func realtimeVoiceDisplayedResponseArtifacts(displayResponse string, bindings []ports.StructuredAgentResponseArtifact) []ports.StructuredAgentResponseArtifact {
	artifacts := make([]ports.StructuredAgentResponseArtifact, 0, min(len(bindings), maxRealtimeVoiceResponseArtifacts))
	seen := map[asset.ID]struct{}{}
	for _, binding := range bindings {
		if len(artifacts) >= maxRealtimeVoiceResponseArtifacts || !containsExactRealtimeVoiceEntityTitle(displayResponse, binding.Title) {
			continue
		}
		if _, duplicate := seen[binding.AssetID]; duplicate {
			continue
		}
		seen[binding.AssetID] = struct{}{}
		artifacts = append(artifacts, binding)
	}
	return artifacts
}

func containsExactRealtimeVoiceEntityTitle(value, title string) bool {
	value = strings.TrimSpace(value)
	title = strings.TrimSpace(title)
	if value == "" || title == "" {
		return false
	}
	pattern := `(^|[^\pL\pN])` + regexp.QuoteMeta(title) + `($|[^\pL\pN])`
	matched, err := regexp.MatchString(pattern, value)
	return err == nil && matched
}

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
			!validContext || context != artifact.Context ||
			!containsExactRealtimeVoiceEntityTitle(displayResponse, artifact.Title) {
			return ports.ErrInvalidProviderInput
		}
		if _, duplicate := seen[artifact.AssetID]; duplicate {
			return ports.ErrInvalidProviderInput
		}
		seen[artifact.AssetID] = struct{}{}
	}
	return nil
}
