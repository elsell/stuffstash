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
	readEvidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "Wireless Microphone", CandidateCount: 2}}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, readEvidence)
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
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill", DestinationPath: []string{"garage"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation}}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"garage-id"}},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"garage-id"}},
	}}
	observations := []agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "garage-id", Title: "Garage", Kind: "location"}}
	readEvidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "garage", CandidateCount: 1}}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, readEvidence); err == nil {
		t.Fatal("expected cross-reference candidate ID to be rejected")
	}
}

func TestRealtimeVoiceInvestigationPolicyTurnsBrokenDestinationChainIntoMissingSuffix(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
		DestinationPath:  []string{"garage", "blue cabinet", "upper shelf"},
		DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer, agentmodel.DestinationKindContainer},
	}
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
	readEvidence := []agentmodel.ReadEvidence{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "drill", CandidateCount: 1},
		{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "garage", CandidateCount: 1},
		{EvidenceRound: 1, ReferenceKey: "destination.1", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "blue cabinet", CandidateCount: 1},
		{EvidenceRound: 1, ReferenceKey: "destination.2", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "upper shelf", CandidateCount: 1},
	}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, readEvidence)
	if err != nil {
		t.Fatalf("canonicalize terminal step: %v", err)
	}
	for index := 2; index < 4; index++ {
		if canonical.Resolutions[index].Status != agentmodel.ResolutionMissing || len(canonical.Resolutions[index].CandidateIDs) != 0 {
			t.Fatalf("expected destination suffix to be missing, got %+v", canonical.Resolutions)
		}
	}
}

func TestRealtimeVoiceInvestigationPolicyRejectsTerminalResolutionWithoutReadCoverage(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
		DestinationPath: []string{"garage", "cabinet"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer},
	}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"drill"}},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"garage"}},
		{ReferenceKey: "destination.1", Status: agentmodel.ResolutionMissing},
	}}
	observations := []agentmodel.CandidateObservation{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill", Title: "Drill", Kind: "item"},
		{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "garage", Title: "Garage", Kind: "location"},
	}
	readEvidence := []agentmodel.ReadEvidence{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "drill", CandidateCount: 1},
		{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "garage", CandidateCount: 1},
	}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, readEvidence); err == nil {
		t.Fatal("expected uncovered missing destination resolution to be rejected")
	}
}

func TestRealtimeVoiceInvestigationPolicyRejectsDestinationKindMutationAcrossTurns(t *testing.T) {
	t.Parallel()

	canonical := agentmodel.Intent{
		Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
		DestinationPath: []string{"Workshop"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
	}
	mutated := canonical
	mutated.DestinationKinds = []agentmodel.DestinationKind{agentmodel.DestinationKindContainer}
	if sameRealtimeVoiceInvestigationIntent(canonical, mutated) {
		t.Fatal("expected destination kinds to be immutable across investigation turns")
	}
}

func TestRealtimeVoiceInvestigationPolicyAcceptsAbsentAfterExecutedZeroMatchSearch(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationExists, SubjectMention: "moon boots"}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAbsent,
	}}}
	readEvidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "moon boots", CandidateCount: 0}}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, nil, readEvidence); err != nil {
		t.Fatalf("expected grounded zero-match absence to be accepted, got %v", err)
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
