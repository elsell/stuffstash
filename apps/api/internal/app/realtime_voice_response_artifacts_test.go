package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceResponseGroundingBindsOnlySelectedAuthorizedEntities(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape:   agentmodel.RequestShapeSingleTarget,
		Kind:           agentmodel.IntentKindRead,
		Operation:      agentmodel.OperationLocate,
		SubjectMention: "cordless drill",
	}
	resolutions := []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject,
		Status:       agentmodel.ResolutionStrong,
		CandidateIDs: []string{"drill"},
	}}
	candidates := map[string]agentmodel.CandidateObservation{
		"drill": {
			EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill",
			Title: "Cordless drill", Kind: "item", ParentAssetID: "toolbox", ParentTitle: "Toolbox", ParentKind: "container",
			ContainmentPath: []string{"Garage", "Toolbox", "Cordless drill"},
		},
		"hidden": {
			EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "hidden",
			Title: "Hidden drill", Kind: "item", ParentAssetID: "hidden-room", ParentTitle: "Hidden room", ParentKind: "location",
		},
	}

	grounding, err := realtimeVoiceInvestigationResponseGrounding(intent, resolutions, candidates)
	if err != nil {
		t.Fatalf("build grounded response: %v", err)
	}
	if len(grounding.Bindings) != 2 {
		t.Fatalf("expected selected asset and immediate parent, got %+v", grounding.Bindings)
	}
	if grounding.Bindings[0].AssetID != asset.ID("drill") || grounding.Bindings[1].AssetID != asset.ID("toolbox") {
		t.Fatalf("unexpected authorized bindings: %+v", grounding.Bindings)
	}
	if grounding.Bindings[0].Context != "Toolbox" || grounding.Bindings[1].Context != "Garage" {
		t.Fatalf("expected authorized containment context, got %+v", grounding.Bindings)
	}
	for _, binding := range grounding.Bindings {
		if binding.AssetID == asset.ID("hidden") || binding.AssetID == asset.ID("hidden-room") {
			t.Fatalf("hidden candidate entered response navigation bindings: %+v", grounding.Bindings)
		}
	}
}

func TestGenerateRealtimeVoiceResponseAttachesBindingsWithoutSendingThemToProvider(t *testing.T) {
	t.Parallel()
	generator := &capturingArtifactResponseGenerator{result: ports.VoiceResponseGenerationResult{
		SpokenResponse: "The drill is in the toolbox.", DisplayResponse: "The Drill is in the Toolbox.",
	}}
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeLocate, Operation: agentmodel.OperationLocate,
		Subject: "drill", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Drill", Kind: "item", ContainmentPath: []string{"Toolbox", "Drill"}}},
	}
	bindings := []ports.StructuredAgentResponseArtifact{
		{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill-id"), Title: "Drill", AssetKind: asset.KindItem},
		{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("toolbox-id"), Title: "Toolbox", AssetKind: asset.KindContainer},
	}
	response, err := (App{}).generateRealtimeVoiceResponse(context.Background(), RealtimeVoiceSession{responseGenerator: generator}, brief, bindings)
	if err != nil {
		t.Fatalf("generate response: %v", err)
	}
	if len(response.Artifacts) != 2 {
		t.Fatalf("expected application-authored artifacts, got %+v", response.Artifacts)
	}
	if generator.input.Brief.Findings[0].Title != "Drill" {
		t.Fatalf("provider did not receive expected presentation brief: %+v", generator.input.Brief)
	}
}

func TestRealtimeVoiceResponseArtifactsRequireExactDisplayedAuthorizedTitles(t *testing.T) {
	t.Parallel()
	bindings := []ports.StructuredAgentResponseArtifact{
		{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill"), Title: "Cordless drill", AssetKind: asset.KindItem},
		{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("toolbox"), Title: "Toolbox", AssetKind: asset.KindContainer},
	}
	artifacts := realtimeVoiceDisplayedResponseArtifacts("The Cordless drill is in the Toolbox.", bindings)
	if len(artifacts) != 2 {
		t.Fatalf("expected both exact displayed references, got %+v", artifacts)
	}
	artifacts = realtimeVoiceDisplayedResponseArtifacts("The drill is nearby.", bindings)
	if len(artifacts) != 0 {
		t.Fatalf("abbreviated or absent titles must not create navigation, got %+v", artifacts)
	}
	artifacts = realtimeVoiceDisplayedResponseArtifacts("Please help us find it.", []ports.StructuredAgentResponseArtifact{{
		Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("us"), Title: "US", AssetKind: asset.KindItem,
	}})
	if len(artifacts) != 0 {
		t.Fatalf("case-distinct ordinary words must not create navigation, got %+v", artifacts)
	}
}

func TestValidateRealtimeVoiceFinalResponseRejectsUnsafeArtifacts(t *testing.T) {
	t.Parallel()
	valid := ports.StructuredAgentResponse{
		Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: "The drill is in the toolbox.", DisplayResponse: "The Drill is in the Toolbox.",
		Artifacts: []ports.StructuredAgentResponseArtifact{{
			Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill"), Title: "Drill", AssetKind: asset.KindItem,
		}},
	}
	if err := validateRealtimeVoiceFinalResponse(valid); err != nil {
		t.Fatalf("expected safe asset reference, got %v", err)
	}
	invalid := []ports.StructuredAgentResponse{
		{Kind: valid.Kind, SpokenResponse: valid.SpokenResponse, DisplayResponse: valid.DisplayResponse, Artifacts: []ports.StructuredAgentResponseArtifact{{Type: ports.StructuredAgentResponseArtifactAssetReference, Title: "Drill", AssetKind: asset.KindItem}}},
		{Kind: valid.Kind, SpokenResponse: valid.SpokenResponse, DisplayResponse: valid.DisplayResponse, Artifacts: []ports.StructuredAgentResponseArtifact{{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill"), Title: "Drill", AssetKind: asset.Kind("unknown")}}},
		{Kind: valid.Kind, SpokenResponse: valid.SpokenResponse, DisplayResponse: valid.DisplayResponse, Artifacts: []ports.StructuredAgentResponseArtifact{{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill"), Title: "Drill", AssetKind: asset.KindItem, Context: " "}}},
		{Kind: valid.Kind, SpokenResponse: valid.SpokenResponse, DisplayResponse: valid.DisplayResponse, Artifacts: []ports.StructuredAgentResponseArtifact{{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill"), Title: "Drill", AssetKind: asset.KindItem}, {Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID("drill"), Title: "Drill again", AssetKind: asset.KindItem}}},
	}
	for _, response := range invalid {
		if err := validateRealtimeVoiceFinalResponse(response); err == nil {
			t.Fatalf("expected unsafe artifact collection to fail: %+v", response.Artifacts)
		}
	}
	overflow := valid
	overflow.Artifacts = make([]ports.StructuredAgentResponseArtifact, maxRealtimeVoiceResponseArtifacts+1)
	for index := range overflow.Artifacts {
		overflow.Artifacts[index] = ports.StructuredAgentResponseArtifact{
			Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: asset.ID(fmt.Sprintf("asset-%d", index)), Title: "Drill", AssetKind: asset.KindItem,
		}
	}
	if err := validateRealtimeVoiceFinalResponse(overflow); err == nil {
		t.Fatalf("expected more than %d artifacts to fail", maxRealtimeVoiceResponseArtifacts)
	}
}

type capturingArtifactResponseGenerator struct {
	input  ports.VoiceResponseGenerationInput
	result ports.VoiceResponseGenerationResult
}

func (g *capturingArtifactResponseGenerator) GenerateResponse(_ context.Context, input ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	g.input = input
	return g.result, nil
}
