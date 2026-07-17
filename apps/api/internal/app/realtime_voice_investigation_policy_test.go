package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceInvestigationPolicyMakesSoleExactTitleDominateDistractors(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Wireless Microphone"}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAmbiguous,
		CandidateIDs: []string{"exact", "partial"},
	}}}
	observations := []agentmodel.CandidateObservation{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "exact", Title: "Wireless Microphone", Kind: "item"},
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "partial", Title: "Wireless Microphone Stand", Kind: "item"},
	}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations)
	if err != nil {
		t.Fatalf("canonicalize terminal step: %v", err)
	}
	resolution := canonical.Resolutions[0]
	if resolution.Status != agentmodel.ResolutionStrong || len(resolution.CandidateIDs) != 1 || resolution.CandidateIDs[0] != "exact" {
		t.Fatalf("expected sole exact title to dominate, got %+v", resolution)
	}
}

func TestRealtimeVoiceInvestigationPolicyRejectsCrossReferenceCandidate(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill", DestinationPath: []string{"garage"}}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"garage-id"}},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"garage-id"}},
	}}
	observations := []agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "garage-id", Title: "Garage", Kind: "location"}}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations); err == nil {
		t.Fatal("expected cross-reference candidate ID to be rejected")
	}
}

func TestRealtimeVoiceInvestigationPolicyTurnsBrokenDestinationChainIntoMissingSuffix(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill", DestinationPath: []string{"garage", "blue cabinet", "upper shelf"}}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"drill"}},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"garage"}},
		{ReferenceKey: "destination.1", Status: agentmodel.ResolutionPlausible, CandidateIDs: []string{"other-cabinet"}},
		{ReferenceKey: "destination.2", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"other-shelf"}},
	}}
	observations := []agentmodel.CandidateObservation{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill", Title: "Drill", Kind: "item"},
		{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "garage", Title: "Garage", Kind: "location"},
		{EvidenceRound: 1, ReferenceKey: "destination.1", CandidateID: "other-cabinet", Title: "Blue cabinet", Kind: "container", ParentAssetID: "basement"},
		{EvidenceRound: 1, ReferenceKey: "destination.2", CandidateID: "other-shelf", Title: "Upper shelf", Kind: "container", ParentAssetID: "other-cabinet"},
	}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations)
	if err != nil {
		t.Fatalf("canonicalize terminal step: %v", err)
	}
	for index := 2; index < 4; index++ {
		if canonical.Resolutions[index].Status != agentmodel.ResolutionMissing || len(canonical.Resolutions[index].CandidateIDs) != 0 {
			t.Fatalf("expected destination suffix to be missing, got %+v", canonical.Resolutions)
		}
	}
}

func TestRealtimeVoiceInvestigationResponseCalibratesPlausibleMatch(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Sarah winter coat"}
	resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionPlausible, CandidateIDs: []string{"clothes"}}}
	candidates := map[string]agentmodel.CandidateObservation{"clothes": {
		EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "clothes", Title: "Sarah Winter Clothes and Shoes", Kind: "container", ContainmentPath: []string{"Basement", "Storage room", "Sarah Winter Clothes and Shoes"},
	}}
	response, err := realtimeVoiceInvestigationResponse(intent, resolutions, candidates)
	if err != nil {
		t.Fatalf("render response: %v", err)
	}
	if response.Kind != ports.StructuredAgentResponseKindAnswer || response.SpokenResponse != "I think you mean Sarah Winter Clothes and Shoes. Its recorded path is Basement / Storage room / Sarah Winter Clothes and Shoes." {
		t.Fatalf("unexpected calibrated response: %+v", response)
	}
}
