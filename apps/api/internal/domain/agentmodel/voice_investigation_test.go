package agentmodel

import (
	"strings"
	"testing"
)

func TestVoiceInvestigationEnumsRejectUnknownValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		valid func(string) bool
	}{
		{name: "intent kind", valid: func(value string) bool { return IntentKind(value).Valid() }},
		{name: "operation", valid: func(value string) bool { return Operation(value).Valid() }},
		{name: "decision", valid: func(value string) bool { return InvestigationDecision(value).Valid() }},
		{name: "read kind", valid: func(value string) bool { return InvestigationReadKind(value).Valid() }},
		{name: "resolution status", valid: func(value string) bool { return ResolutionStatus(value).Valid() }},
		{name: "phase", valid: func(value string) bool { return InvestigationPhase(value).Valid() }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.valid("") || test.valid("provider_specific") {
				t.Fatalf("expected %s to reject empty and unknown values", test.name)
			}
		})
	}
}

func TestSemanticReferenceKeyAcceptsOnlyCanonicalSubjectAndDestinationKeys(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"subject", "destination.0", "destination.5"} {
		key, ok := NewSemanticReferenceKey(value)
		if !ok || key.String() != value {
			t.Fatalf("expected %q to be a valid semantic reference, got %q, %t", value, key, ok)
		}
	}
	for _, value := range []string{"", "source", "destination", "destination.-1", "destination.00", "destination.6"} {
		if _, ok := NewSemanticReferenceKey(value); ok {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestIntentValidatesOperationCoherenceAndBounds(t *testing.T) {
	t.Parallel()

	valid := Intent{Kind: IntentKindChange, Operation: OperationMove, SubjectMention: "drill", DestinationPath: []string{"garage", "cabinet"}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid move intent, got %v", err)
	}

	tests := []struct {
		name   string
		intent Intent
	}{
		{name: "read kind with write operation", intent: Intent{Kind: IntentKindRead, Operation: OperationMove, SubjectMention: "drill"}},
		{name: "move without destination", intent: Intent{Kind: IntentKindChange, Operation: OperationMove, SubjectMention: "drill"}},
		{name: "create without subject", intent: Intent{Kind: IntentKindChange, Operation: OperationCreate, NewAssetKind: "item"}},
		{name: "unsupported asset kind", intent: Intent{Kind: IntentKindChange, Operation: OperationCreate, SubjectMention: "charger", NewAssetKind: "vehicle"}},
		{name: "too many destination segments", intent: Intent{Kind: IntentKindChange, Operation: OperationCreate, SubjectMention: "charger", DestinationPath: []string{"a", "b", "c", "d", "e", "f", "g"}}},
		{name: "unbounded details", intent: Intent{Kind: IntentKindChange, Operation: OperationCheckout, SubjectMention: "drill", Details: strings.Repeat("x", MaxInvestigationDetailRunes+1)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.intent.Validate(); err == nil {
				t.Fatalf("expected invalid intent %+v", test.intent)
			}
		})
	}
}

func TestSearchRequestValidatesReadSpecificShapeAndProbeBounds(t *testing.T) {
	t.Parallel()

	subject := SemanticReferenceSubject
	search := SearchRequest{
		ReferenceKey: subject,
		ReadKind:     InvestigationReadSearchAssets,
		Mention:      "Sarah winter coat",
		SearchProbes: []string{"Sarah winter coat", "Sarah winter clothes", "winter clothes and shoes"},
	}
	if err := search.Validate(); err != nil {
		t.Fatalf("expected valid fuzzy search request, got %v", err)
	}
	if err := (SearchRequest{ReferenceKey: subject, ReadKind: InvestigationReadListInventory}).Validate(); err != nil {
		t.Fatalf("expected inventory list without probes to be valid, got %v", err)
	}
	if err := (SearchRequest{ReferenceKey: subject, ReadKind: InvestigationReadAssetDetail, VisibleAssetID: "asset-1"}).Validate(); err != nil {
		t.Fatalf("expected grounded detail read to be valid, got %v", err)
	}

	tests := []SearchRequest{
		{ReferenceKey: subject, ReadKind: InvestigationReadSearchAssets, Mention: "coat"},
		{ReferenceKey: subject, ReadKind: InvestigationReadAssetDetail},
		{ReferenceKey: subject, ReadKind: InvestigationReadSearchAssets, Mention: "coat", SearchProbes: []string{"coat", " COAT "}},
		{ReferenceKey: subject, ReadKind: InvestigationReadSearchAssets, Mention: "coat", SearchProbes: make([]string, MaxSearchProbesPerRequest+1)},
		{ReferenceKey: "other", ReadKind: InvestigationReadSearchAssets, Mention: "coat", SearchProbes: []string{"coat"}},
	}
	for _, request := range tests {
		if err := request.Validate(); err == nil {
			t.Fatalf("expected invalid search request %+v", request)
		}
	}
}

func TestCandidateObservationBoundsAuthorizedEvidenceShape(t *testing.T) {
	t.Parallel()

	observation := CandidateObservation{
		EvidenceRound: 1,
		ReferenceKey:  SemanticReferenceSubject,
		CandidateID:   "asset-1",
		Title:         "Sarah Winter Clothes and Shoes",
		Kind:          "container",
		ContainmentPath: []string{
			"Basement", "Storage room", "Shelf 2",
		},
		MatchedProbes: []string{"Sarah winter clothes"},
		Facts:         []string{"active", "inside Shelf 2"},
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("expected valid candidate observation, got %v", err)
	}

	observation.CandidateID = ""
	if err := observation.Validate(); err == nil {
		t.Fatal("expected candidate observation without an authorized candidate ID to fail")
	}
	observation.CandidateID = "asset-1"
	observation.Facts = make([]string, MaxObservationFacts+1)
	if err := observation.Validate(); err == nil {
		t.Fatal("expected unbounded candidate facts to fail")
	}
}

func TestResolutionEnforcesStatusCandidateCardinality(t *testing.T) {
	t.Parallel()

	valid := []Resolution{
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionStrong, CandidateIDs: []string{"asset-1"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionPlausible, CandidateIDs: []string{"asset-1"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionAmbiguous, CandidateIDs: []string{"asset-1", "asset-2"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionCollection, CandidateIDs: []string{"asset-1", "asset-2"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionAbsent},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionMissing},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionUnsupported},
	}
	for _, resolution := range valid {
		if err := resolution.Validate(); err != nil {
			t.Fatalf("expected valid resolution %+v, got %v", resolution, err)
		}
	}

	invalid := []Resolution{
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionStrong},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionPlausible, CandidateIDs: []string{"asset-1", "asset-2"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionAmbiguous, CandidateIDs: []string{"asset-1"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionAbsent, CandidateIDs: []string{"asset-1"}},
		{ReferenceKey: SemanticReferenceSubject, Status: ResolutionCollection, CandidateIDs: []string{"asset-1", "asset-1"}},
	}
	for _, resolution := range invalid {
		if err := resolution.Validate(); err == nil {
			t.Fatalf("expected invalid resolution %+v", resolution)
		}
	}
}

func TestInvestigationInputValidatesPhaseRoundAndEvidenceBounds(t *testing.T) {
	t.Parallel()

	intent := Intent{Kind: IntentKindRead, Operation: OperationLocate, SubjectMention: "winter clothes"}
	input := InvestigationInput{
		Phase:             InvestigationPhaseEvidenceAssessment,
		PromptVersion:     "voice-investigation-v1",
		SchemaVersion:     "voice-investigation-v1",
		Transcript:        "Where are Sarah's winter coat?",
		EvidenceRound:     1,
		MaxEvidenceRounds: MaxEvidenceRounds,
		CanonicalIntent:   &intent,
		PreviousRequests: []SearchRequest{{
			ReferenceKey: SemanticReferenceSubject,
			ReadKind:     InvestigationReadSearchAssets,
			Mention:      "Sarah winter coat",
			SearchProbes: []string{"Sarah winter coat", "Sarah winter clothes"},
		}},
		Observations: []CandidateObservation{{EvidenceRound: 1, ReferenceKey: SemanticReferenceSubject, CandidateID: "asset-1", Title: "Sarah Winter Clothes and Shoes", Kind: "container"}},
	}
	if err := input.Validate(); err != nil {
		t.Fatalf("expected valid evidence input, got %v", err)
	}

	input.MaxEvidenceRounds = MaxEvidenceRounds + 1
	if err := input.Validate(); err == nil {
		t.Fatal("expected evidence budget above the production maximum to fail")
	}
	input.MaxEvidenceRounds = MaxEvidenceRounds
	input.Phase = InvestigationPhaseInitial
	if err := input.Validate(); err == nil {
		t.Fatal("expected initial phase with prior evidence to fail")
	}
}

func TestInvestigationStepValidatesDecisionSpecificPayload(t *testing.T) {
	t.Parallel()

	intent := Intent{Kind: IntentKindRead, Operation: OperationLocate, SubjectMention: "winter clothes"}
	request := SearchRequest{ReferenceKey: SemanticReferenceSubject, ReadKind: InvestigationReadSearchAssets, Mention: "winter clothes", SearchProbes: []string{"winter clothes"}}
	resolution := Resolution{ReferenceKey: SemanticReferenceSubject, Status: ResolutionStrong, CandidateIDs: []string{"asset-1"}}

	steps := []InvestigationStep{
		{Decision: InvestigationDecisionSearch, Intent: intent, SearchRequests: []SearchRequest{request}},
		{Decision: InvestigationDecisionSearchAgain, Intent: intent, SearchRequests: []SearchRequest{request}},
		{Decision: InvestigationDecisionFinish, Intent: intent, Resolutions: []Resolution{resolution}},
	}
	for _, step := range steps {
		if err := step.Validate(); err != nil {
			t.Fatalf("expected valid investigation step %+v, got %v", step, err)
		}
	}

	invalid := []InvestigationStep{
		{Decision: InvestigationDecisionSearch, Intent: intent},
		{Decision: InvestigationDecisionSearchAgain, Intent: intent, SearchRequests: []SearchRequest{request}, Resolutions: []Resolution{resolution}},
		{Decision: InvestigationDecisionFinish, Intent: intent, SearchRequests: []SearchRequest{request}, Resolutions: []Resolution{resolution}},
		{Decision: InvestigationDecisionFinish, Intent: intent},
		{Decision: InvestigationDecisionFinish, Intent: intent, Resolutions: []Resolution{resolution, resolution}},
	}
	for _, step := range invalid {
		if err := step.Validate(); err == nil {
			t.Fatalf("expected invalid investigation step %+v", step)
		}
	}
}
