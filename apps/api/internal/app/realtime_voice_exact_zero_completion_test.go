package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func TestRealtimeVoiceExactOrZeroCompletionOverridesFinishIntentDrift(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate,
		SubjectMention: "Power Adapter", NewAssetKind: "item", DestinationPath: []string{"Pedal Case"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindContainer},
	}
	drifted := intent
	drifted.DestinationPath = []string{"Rehearsal Room", "Pedal Case"}
	drifted.DestinationKinds = []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: drifted, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionMissing},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"pedal-case"}},
	}}
	completed, ok := realtimeVoiceExactOrZeroCompletion(intent, step,
		[]agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "pedal-case", Title: "Pedal Case", Kind: "container"}},
		[]agentmodel.ReadEvidence{
			{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "power adapter", CandidateCount: 0},
			{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "pedal case", CandidateCount: 1},
		},
	)
	if !ok || !sameRealtimeVoiceInvestigationIntent(completed.Intent, intent) || len(completed.Resolutions) != 2 {
		t.Fatalf("expected deterministic original-intent completion, got ok=%t step=%+v", ok, completed)
	}
}

func TestRealtimeVoiceExactOrZeroCompletionIgnoresWrongKindDestinationDistractors(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate,
		SubjectMention: "Label Cable", NewAssetKind: "item", DestinationPath: []string{"Scanner Cart"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindContainer},
	}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearchAgain, Intent: intent}
	completed, ok := realtimeVoiceExactOrZeroCompletion(intent, step,
		[]agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "barcode-scanner", Title: "Barcode Scanner", Kind: "item"}},
		[]agentmodel.ReadEvidence{
			{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "label cable", CandidateCount: 0},
			{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "scanner cart", CandidateCount: 1},
		},
	)
	if !ok || len(completed.Resolutions) != 2 || completed.Resolutions[1].Status != agentmodel.ResolutionMissing {
		t.Fatalf("expected wrong-kind destination distractor to count as missing, got ok=%t step=%+v", ok, completed)
	}
}
