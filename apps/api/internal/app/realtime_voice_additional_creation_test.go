package app

import (
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func TestAdditionalCreationDoesNotReuseExistingSubject(t *testing.T) {
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate, SubjectMention: "Charger", NewAssetKind: "item", CreationMode: agentmodel.CreationModeAdditional, CreationEvidence: "I bought another charger"}
	candidate := agentmodel.CandidateObservation{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "existing-charger", Title: "Charger", Kind: "item", LifecycleState: "active"}
	resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{candidate.CandidateID}}}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: resolutions}
	evidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "charger", CandidateCount: 1}}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, []agentmodel.CandidateObservation{candidate}, evidence)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := compileRealtimeVoiceActionPlan(intent, canonical.Resolutions, map[string]agentmodel.CandidateObservation{candidate.CandidateID: candidate})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Disposition != realtimeVoicePlanReady || len(plan.Commands) != 1 || !strings.Contains(plan.ConfirmationSummary, "additional") {
		t.Fatalf("additional item not reviewable: %+v", plan)
	}
	if _, ok := plan.Commands[0].Arguments["assetId"]; ok {
		t.Fatal("new physical item reused existing identity")
	}
	intent.CreationMode = agentmodel.CreationModeRecord
	intent.CreationEvidence = ""
	plan, err = compileRealtimeVoiceActionPlan(intent, resolutions, map[string]agentmodel.CandidateObservation{candidate.CandidateID: candidate})
	if err != nil || plan.Disposition != realtimeVoicePlanNoOp {
		t.Fatalf("ordinary recording lost duplicate protection: %+v %v", plan, err)
	}
}
