package app

import (
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxRealtimeVoiceResponseArtifacts = tools.MaxRealtimeVoiceResponseArtifacts

func validateRealtimeVoiceResponseArtifacts(displayResponse string, artifacts []ports.StructuredAgentResponseArtifact) error {
	return tools.ValidateRealtimeVoiceResponseArtifacts(displayResponse, artifacts)
}
