package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func TestRealtimeVoiceInvestigationPolicyMakesSoleExactTitleDominateDistractors(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Wireless Microphone"}
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

func TestRealtimeVoiceInvestigationPolicyDowngradesNonExactStrongMatch(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "tools"}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"toolbox"},
	}}}
	observations := []agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "toolbox", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "Toolbox"}}}
	evidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "tools", CandidateCount: 1}}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, evidence)
	if err != nil {
		t.Fatalf("canonicalize terminal step: %v", err)
	}
	if canonical.Resolutions[0].Status != agentmodel.ResolutionPlausible {
		t.Fatalf("expected non-exact strong match to be calibrated as plausible, got %+v", canonical.Resolutions[0])
	}
}

func TestRealtimeVoiceInvestigationPolicyRejectsCrossReferenceCandidate(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill", DestinationPath: []string{"garage"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation}}
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
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
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

func TestRealtimeVoiceInvestigationPolicyRequiresClarificationForPlausibleWriteDestination(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
		DestinationPath: []string{"Live Room"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
	}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"drill"}},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionPlausible, CandidateIDs: []string{"rehearsal-room"}},
	}}
	observations := []agentmodel.CandidateObservation{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill", Title: "Drill", Kind: "item"},
		{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "rehearsal-room", Title: "Rehearsal Room", Kind: "location"},
	}
	readEvidence := []agentmodel.ReadEvidence{
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "drill", CandidateCount: 1},
		{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "live room", CandidateCount: 1},
	}
	canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, readEvidence)
	if err != nil {
		t.Fatalf("canonicalize terminal step: %v", err)
	}
	brief, err := realtimeVoiceInvestigationResponseBrief(canonical.Intent, canonical.Resolutions, map[string]agentmodel.CandidateObservation{
		"drill": observations[0], "rehearsal-room": observations[1],
	})
	if err != nil {
		t.Fatalf("build clarification: %v", err)
	}
	if brief.Kind != agentmodel.ResponseBriefKindClarification || brief.Subject != "Live Room" || len(brief.Findings) != 1 || brief.Findings[0].Title != "Rehearsal Room" {
		t.Fatalf("expected plausible write destination to require explicit clarification, got %+v", brief)
	}
}

func TestRealtimeVoiceInvestigationPolicyRejectsTerminalResolutionWithoutReadCoverage(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
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
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
		DestinationPath: []string{"Workshop"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
	}
	mutated := canonical
	mutated.DestinationKinds = []agentmodel.DestinationKind{agentmodel.DestinationKindContainer}
	if sameRealtimeVoiceInvestigationIntent(canonical, mutated) {
		t.Fatal("expected destination kinds to be immutable across investigation turns")
	}
}

func TestRealtimeVoiceInvestigationPolicyAnchorsInitialDetailsAcrossEvidenceTurns(t *testing.T) {
	t.Parallel()
	canonical := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationCheckoutStatus, SubjectMention: "loaner flashlight", Details: ""}
	provider := canonical
	provider.Details = "Checked out at a provider-observed time"
	if !sameRealtimeVoiceInvestigationIntent(canonical, provider) {
		t.Fatal("evidence-only detail rewrite should not mutate anchored intent identity")
	}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: provider, Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"flashlight-1"}}}}
	observations := []agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "flashlight-1", Title: "Loaner flashlight", Kind: "item", LifecycleState: "active"}}
	evidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "loaner flashlight", CandidateCount: 1, LifecycleScope: agentmodel.LifecycleScopeActive}}
	result, err := canonicalRealtimeVoiceInvestigationStep(canonical, step, observations, evidence)
	if err != nil {
		t.Fatalf("canonicalize evidence detail rewrite: %v", err)
	}
	if result.Intent.Details != "" {
		t.Fatalf("evidence turn changed anchored details: %+v", result.Intent)
	}
}

func TestRealtimeVoiceInvestigationPolicyAcceptsAbsentAfterExecutedZeroMatchSearch(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationExists, SubjectMention: "moon boots"}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAbsent,
	}}}
	readEvidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "moon boots", CandidateCount: 0}}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, nil, readEvidence); err != nil {
		t.Fatalf("expected grounded zero-match absence to be accepted, got %v", err)
	}
}

func TestRealtimeVoiceInvestigationPolicyDerivesNoCandidateStatusFromReferenceRole(t *testing.T) {
	t.Parallel()

	t.Run("missing existing source becomes absent", func(t *testing.T) {
		t.Parallel()
		intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationArchive, SubjectMention: "drill"}
		step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
			ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionMissing,
		}}}
		evidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "drill", CandidateCount: 0}}
		canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, nil, evidence)
		if err != nil {
			t.Fatalf("canonicalize source status: %v", err)
		}
		if canonical.Resolutions[0].Status != agentmodel.ResolutionAbsent {
			t.Fatalf("expected existing source absence, got %+v", canonical.Resolutions)
		}
	})

	t.Run("absent create subject and destination become missing", func(t *testing.T) {
		t.Parallel()
		intent := agentmodel.Intent{
			RequestShape: agentmodel.RequestShapeSingleTarget,
			Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate, SubjectMention: "charger", NewAssetKind: "item",
			DestinationPath: []string{"garage"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
		}
		step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
			{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAbsent},
			{ReferenceKey: "destination.0", Status: agentmodel.ResolutionAbsent},
		}}
		evidence := []agentmodel.ReadEvidence{
			{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "charger", CandidateCount: 0},
			{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "garage", CandidateCount: 0},
		}
		canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, nil, evidence)
		if err != nil {
			t.Fatalf("canonicalize create statuses: %v", err)
		}
		for _, resolution := range canonical.Resolutions {
			if resolution.Status != agentmodel.ResolutionMissing {
				t.Fatalf("expected create references to be missing, got %+v", canonical.Resolutions)
			}
		}
	})

	t.Run("absent move destination becomes missing", func(t *testing.T) {
		t.Parallel()
		intent := agentmodel.Intent{
			RequestShape: agentmodel.RequestShapeSingleTarget,
			Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "drill",
			DestinationPath: []string{"kitchen"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
		}
		step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
			{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"drill-1"}},
			{ReferenceKey: "destination.0", Status: agentmodel.ResolutionAbsent},
		}}
		observations := []agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill-1", Title: "Drill", Kind: "item"}}
		evidence := []agentmodel.ReadEvidence{
			{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "drill", CandidateCount: 1},
			{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "kitchen", CandidateCount: 0},
		}
		canonical, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, evidence)
		if err != nil {
			t.Fatalf("canonicalize destination status: %v", err)
		}
		if canonical.Resolutions[1].Status != agentmodel.ResolutionMissing {
			t.Fatalf("expected missing destination, got %+v", canonical.Resolutions)
		}
	})
}

func TestRealtimeVoiceInvestigationResponseCalibratesPlausibleMatch(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Sarah winter coat"}
	resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionPlausible, CandidateIDs: []string{"clothes"}}}
	candidates := map[string]agentmodel.CandidateObservation{"clothes": {
		EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "clothes", Title: "Sarah Winter Clothes and Shoes", Kind: "container", ContainmentPath: []string{"Basement", "Storage room", "Sarah Winter Clothes and Shoes"},
	}}
	brief, err := realtimeVoiceInvestigationResponseBrief(intent, resolutions, candidates)
	if err != nil {
		t.Fatalf("render response: %v", err)
	}
	if brief.Kind != agentmodel.ResponseBriefKindAnswer || brief.Confidence != agentmodel.ResponseConfidencePlausible || brief.Mode != agentmodel.ResponseAnswerModeLocate {
		t.Fatalf("unexpected calibrated brief: %+v", brief)
	}
}

func TestRealtimeVoiceInvestigationResponseAnswersWhereCollectionIs(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeCollectionTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "tools"}
	resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection, CandidateIDs: []string{"toolbox"}}}
	candidates := map[string]agentmodel.CandidateObservation{"toolbox": {
		EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "toolbox", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "Toolbox"},
	}}
	brief, err := realtimeVoiceInvestigationResponseBrief(intent, resolutions, candidates)
	if err != nil {
		t.Fatalf("render response: %v", err)
	}
	if brief.Mode != agentmodel.ResponseAnswerModeLocate || brief.Confidence != agentmodel.ResponseConfidencePlausible || len(brief.Findings) != 1 || brief.Findings[0].Title != "Toolbox" {
		t.Fatalf("expected grounded collection location brief, got %+v", brief)
	}
}

func TestRealtimeVoiceInvestigationResponsePreservesDifferentCollectionLocations(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeCollectionTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "tools"}
	resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection, CandidateIDs: []string{"drill", "shears"}}}
	candidates := map[string]agentmodel.CandidateObservation{
		"drill":  {EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill", Title: "Cordless drill", Kind: "item", ContainmentPath: []string{"Garage", "Toolbox", "Cordless drill"}},
		"shears": {EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "shears", Title: "Garden shears", Kind: "item", ContainmentPath: []string{"Garage", "Garden shears"}},
	}
	brief, err := realtimeVoiceInvestigationResponseBrief(intent, resolutions, candidates)
	if err != nil {
		t.Fatalf("render response: %v", err)
	}
	if len(brief.Findings) != 2 || len(brief.Findings[0].ContainmentPath) != 2 || len(brief.Findings[1].ContainmentPath) != 2 {
		t.Fatalf("expected each location to survive the brief, got %+v", brief)
	}
}

func TestRealtimeVoiceInvestigationResponseUsesHouseholdLanguageForLists(t *testing.T) {
	t.Parallel()
	candidates := map[string]agentmodel.CandidateObservation{
		"bottle": {EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "bottle", Title: "Water bottle", Kind: "item"},
		"laptop": {EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "laptop", Title: "Laptop", Kind: "item"},
	}
	resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection, CandidateIDs: []string{"bottle", "laptop"}}}

	inventory, err := realtimeVoiceInvestigationResponseBrief(agentmodel.Intent{RequestShape: agentmodel.RequestShapeCollectionTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationListInventory, SubjectMention: "items"}, resolutions, candidates)
	if err != nil {
		t.Fatalf("render inventory response: %v", err)
	}
	if inventory.Mode != agentmodel.ResponseAnswerModeInventory || len(inventory.Findings) != 2 {
		t.Fatalf("unexpected inventory response: %+v", inventory)
	}

	contentsResolution := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection, CandidateIDs: []string{"bottle"}}}
	contents, err := realtimeVoiceInvestigationResponseBrief(agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationListContents, SubjectMention: "Office"}, contentsResolution, map[string]agentmodel.CandidateObservation{"bottle": candidates["bottle"]})
	if err != nil {
		t.Fatalf("render contents response: %v", err)
	}
	if contents.Mode != agentmodel.ResponseAnswerModeContents || contents.Subject != "Office" || len(contents.Findings) != 1 {
		t.Fatalf("unexpected contents response: %+v", contents)
	}
}

func TestRealtimeVoiceInvestigationResponseNamesMissingExistingSubject(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "my passport", DestinationPath: []string{"Office"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation}}
	brief, err := realtimeVoiceInvestigationResponseBrief(intent, []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAbsent}}, nil)
	if err != nil {
		t.Fatalf("render absent source: %v", err)
	}
	if brief.Kind != agentmodel.ResponseBriefKindClarification || brief.Subject != "my passport" || brief.Confidence != agentmodel.ResponseConfidenceAbsent {
		t.Fatalf("expected useful subject-specific clarification, got %+v", brief)
	}
}

func TestRealtimeVoiceResponseFindingsAreBoundedForSpokenRealization(t *testing.T) {
	t.Parallel()
	ids := make([]string, 0, 20)
	candidates := map[string]agentmodel.CandidateObservation{}
	for index := 0; index < 20; index++ {
		id := fmt.Sprintf("item-%d", index)
		title := fmt.Sprintf("Household item with a deliberately descriptive title %d", index)
		ids = append(ids, id)
		candidates[id] = agentmodel.CandidateObservation{
			EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: id,
			Title: title, Kind: "item", LifecycleState: "active", CheckoutState: "available",
			ContainmentPath: []string{"A very descriptive room", "A very descriptive cabinet", title},
			Facts:           []string{strings.Repeat("history ", 30), "Moved to the cabinet"},
		}
	}
	findings, truncated, err := realtimeVoiceResponseFindings(agentmodel.ResponseAnswerModeInventory, ids, candidates)
	if err != nil {
		t.Fatalf("bound response findings: %v", err)
	}
	if !truncated || len(findings) == 0 || len(findings) >= len(ids) {
		t.Fatalf("expected a non-empty truncated presentation subset, got truncated=%t count=%d", truncated, len(findings))
	}
	for _, finding := range findings {
		if len(finding.ContainmentPath) > 2 || len(finding.Facts) > 3 || !finding.FactsTruncated {
			t.Fatalf("expected bounded path and facts, got %+v", finding)
		}
	}
}

func TestRealtimeVoiceInventoryBriefPrioritizesItemsOverPlaces(t *testing.T) {
	t.Parallel()
	ids := make([]string, 0, 9)
	candidates := map[string]agentmodel.CandidateObservation{}
	for index := 0; index < 8; index++ {
		id := fmt.Sprintf("place-%d", index)
		ids = append(ids, id)
		candidates[id] = agentmodel.CandidateObservation{
			EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: id,
			Title: fmt.Sprintf("A location %d", index), Kind: "location",
		}
	}
	ids = append(ids, "item")
	candidates["item"] = agentmodel.CandidateObservation{
		EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "item",
		Title: "Zebra label maker", Kind: "item",
	}
	brief, err := realtimeVoiceInvestigationResponseBrief(
		agentmodel.Intent{RequestShape: agentmodel.RequestShapeCollectionTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationListInventory, SubjectMention: "stuff"},
		[]agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection, CandidateIDs: ids}}, candidates,
	)
	if err != nil {
		t.Fatalf("build inventory brief: %v", err)
	}
	if len(realtimeVoiceInventoryResponseFindings(brief.Findings)) != 1 || realtimeVoiceInventoryResponseFindings(brief.Findings)[0].Title != "Zebra label maker" {
		t.Fatalf("expected bounded inventory summary to retain the item, got %+v", brief.Findings)
	}
}

func TestRealtimeVoiceResponseFindingRetainsEvidenceAfterOversizedFact(t *testing.T) {
	t.Parallel()
	finding := realtimeVoiceResponseFinding("finding.0", agentmodel.CandidateObservation{
		EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill", Title: "Drill", Kind: "item",
		Facts: []string{strings.Repeat("oversized history detail ", 20), "Returned July 3"},
	})
	if len(finding.Facts) == 0 {
		t.Fatal("expected at least one bounded history fact")
	}
	if finding.Facts[0] != "Returned July 3" {
		t.Fatalf("expected later bounded fact to survive, got %+v", finding.Facts)
	}
	if !finding.FactsTruncated {
		t.Fatal("expected omitted oversized evidence to be disclosed")
	}

	onlyOversized := realtimeVoiceResponseFinding("finding.0", agentmodel.CandidateObservation{
		EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "case", Title: "Camera case", Kind: "item",
		Facts: []string{strings.Repeat("checked out and returned with detailed notes ", 20)},
	})
	if len(onlyOversized.Facts) != 1 || len(onlyOversized.Facts[0]) > 180 || !onlyOversized.FactsTruncated {
		t.Fatalf("expected a safe bounded fallback fact, got %+v", onlyOversized)
	}
}
