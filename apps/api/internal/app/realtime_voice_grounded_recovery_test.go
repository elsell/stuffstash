package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestGroundedRecoveryHonorsCancellationAndTextSafety(t *testing.T) {
	brief := agentmodel.GroundedVoiceResponseBrief{Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeLocate, Operation: agentmodel.OperationLocate, Subject: "clothes", Confidence: agentmodel.ResponseConfidenceStrong, Findings: []agentmodel.ResponseFinding{{FactKey: "clothes", Title: "Baby clothes", Kind: "item", ContainmentPath: []string{"Attic", "Baby clothes"}}}}
	response, err := (App{}).generateRealtimeVoiceResponse(context.Background(), RealtimeVoiceSession{}, brief, nil)
	if err != nil || response.SpokenResponse != "Baby clothes is in Attic." {
		t.Fatalf("missing provider recovery: %+v %v", response, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (App{}).recoverGroundedVoiceResponse(ctx, brief, nil, "generation_failed"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled recovery: %v", err)
	}
	brief.Findings[0].Title = "search_authorized_assets"
	if _, err := (App{}).recoverGroundedVoiceResponse(context.Background(), brief, nil, "invalid_wording"); !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("unsafe fallback accepted: %v", err)
	}
}
