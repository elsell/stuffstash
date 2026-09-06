package app

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

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
