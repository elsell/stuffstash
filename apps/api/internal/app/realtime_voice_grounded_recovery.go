package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type groundedResponseFallbackReason string

const (
	groundedResponseGenerationFailed groundedResponseFallbackReason = "generation_failed"
	groundedResponseInvalidWording   groundedResponseFallbackReason = "invalid_wording"
)

func (a App) recoverGroundedVoiceResponse(ctx context.Context, brief agentmodel.GroundedVoiceResponseBrief, bindings []ports.StructuredAgentResponseArtifact, reason groundedResponseFallbackReason) (ports.StructuredAgentResponse, error) {
	if err := ctx.Err(); err != nil {
		return ports.StructuredAgentResponse{}, err
	}
	spoken, err := agentmodelapp.RenderGroundedVoiceResponse(brief, 500)
	if err != nil {
		return ports.StructuredAgentResponse{}, err
	}
	displayed, err := agentmodelapp.RenderGroundedVoiceResponse(brief, 1000)
	if err != nil {
		return ports.StructuredAgentResponse{}, err
	}
	response := ports.StructuredAgentResponse{Kind: realtimeVoiceStructuredResponseKind(brief.Kind), SpokenResponse: spoken, DisplayResponse: displayed, Artifacts: realtimeVoiceDisplayedResponseArtifacts(displayed, bindings)}
	if err := validateRealtimeVoiceFinalResponse(response); err != nil {
		return ports.StructuredAgentResponse{}, err
	}
	if a.observer != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventRealtimeVoiceResponseFallback, Message: "voice response rendered from grounded evidence", Fields: map[string]string{"reason": string(reason)}})
	}
	return response, nil
}
