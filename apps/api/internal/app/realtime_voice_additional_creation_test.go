package app

import (
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
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

func TestAdditionalCreationEvidenceComesOnlyFromUserWords(t *testing.T) {
	intent := agentmodel.Intent{CreationMode: agentmodel.CreationModeAdditional, CreationEvidence: "I bought another charger"}
	if realtimeVoiceCreationEvidenceGrounded(intent, "Record my charger", nil) {
		t.Fatal("invented additional intent accepted")
	}
	turns := []ports.AgentConversationTurn{{Role: ports.AgentConversationRoleAssistant, Text: "I bought another charger"}}
	if realtimeVoiceCreationEvidenceGrounded(intent, "Record my charger", turns) {
		t.Fatal("assistant supplied creation authority")
	}
	turns[0].Role = ports.AgentConversationRoleUser
	if !realtimeVoiceCreationEvidenceGrounded(intent, "Put it on the shelf", turns) {
		t.Fatal("user's retained acquisition statement lost")
	}
	if !realtimeVoiceCreationEvidenceGrounded(intent, "I bought another charger. Put it away.", nil) {
		t.Fatal("literal user evidence lost")
	}
}

func TestAdditionalCreationWithMultipleMatchesStillRequiresDiscovery(t *testing.T) {
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate, SubjectMention: "Charger", NewAssetKind: "item", CreationMode: agentmodel.CreationModeAdditional, CreationEvidence: "I bought another charger"}
	candidates := []agentmodel.CandidateObservation{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "first", Title: "Charger", Kind: "item", LifecycleState: "active"},
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "second", Title: "Charger", Kind: "item", LifecycleState: "active"},
	}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAmbiguous, CandidateIDs: []string{"first", "second"}}}}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, candidates, nil); err == nil {
		t.Fatal("additional creation bypassed authorized discovery")
	}
	evidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "charger", CandidateCount: 2}}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, candidates, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if len(canonical.Resolutions) != 1 || canonical.Resolutions[0].Status != agentmodel.ResolutionMissing || len(canonical.Resolutions[0].CandidateIDs) != 0 {
		t.Fatalf("existing copies prevented a new physical instance: %+v", canonical.Resolutions)
	}
	plan, err := compileRealtimeVoiceActionPlan(intent, canonical.Resolutions, map[string]agentmodel.CandidateObservation{"first": candidates[0], "second": candidates[1]})
	if err != nil || plan.Disposition != realtimeVoicePlanReady || len(plan.Commands) != 1 {
		t.Fatalf("additional instance not proposed: %+v %v", plan, err)
	}
	if _, exists := plan.Commands[0].Arguments["assetId"]; exists {
		t.Fatal("additional creation reused an existing copy")
	}
}
