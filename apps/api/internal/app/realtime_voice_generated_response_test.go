package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestValidateRealtimeVoiceGeneratedResponseRequiresGroundedLocation(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeLocate, Operation: agentmodel.OperationLocate,
		Subject: "tools", Confidence: agentmodel.ResponseConfidencePlausible,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "Toolbox"}}},
	}
	valid := ports.VoiceResponseGenerationResult{SpokenResponse: "Your tools are probably in the toolbox in the garage.", DisplayResponse: "Your tools are probably in the Toolbox in the Garage."}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected grounded natural response, got %v", err)
	}

	invalid := []ports.VoiceResponseGenerationResult{
		{SpokenResponse: "I found one visible match: Toolbox.", DisplayResponse: "I found one visible match: Toolbox."},
		{SpokenResponse: "I found your tools.", DisplayResponse: "I found your tools."},
		{SpokenResponse: "Your tools are in the basement.", DisplayResponse: "Your tools are in the basement."},
		{SpokenResponse: "Your tools are in the toolbox. Is that okay?", DisplayResponse: "Your tools are in the Toolbox. Is that okay?"},
		{SpokenResponse: "Your tools are in the toolbox.", DisplayResponse: "Your tools are in the Toolbox."},
		{SpokenResponse: "I think the toolbox might be in the garage.", DisplayResponse: "I think the Toolbox might be in the Garage."},
		{SpokenResponse: "Your passport is in the basement.", DisplayResponse: "Your tools are probably in the Toolbox in the Garage."},
	}
	for _, result := range invalid {
		if err := validateRealtimeVoiceGeneratedResponse(brief, result); err == nil {
			t.Fatalf("expected invalid generated response: %+v", result)
		}
	}
}

func TestValidateRealtimeVoiceGeneratedResponseRestrictsQuestionsToClarifications(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: agentmodel.OperationLocate,
		Subject: "drill", Confidence: agentmodel.ResponseConfidenceAmbiguous,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Blue drill", Kind: "item"}, {FactKey: "finding.1", Title: "Red drill", Kind: "item"}},
	}
	result := ports.VoiceResponseGenerationResult{SpokenResponse: "Did you mean the blue drill or the red drill?", DisplayResponse: "Did you mean the Blue drill or the Red drill?"}
	if err := validateRealtimeVoiceGeneratedResponse(brief, result); err != nil {
		t.Fatalf("expected grounded clarification, got %v", err)
	}
}

func TestValidateRealtimeVoiceGeneratedResponseAllowsUnassignedLocation(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeLocate, Operation: agentmodel.OperationLocate,
		Subject: "camp medicine", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Camp medicine", Kind: "item"}},
	}
	result := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "I found Camp medicine, but it isn't assigned to a location.",
		DisplayResponse: "I found Camp medicine, but it isn't assigned to a location.",
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, result); err != nil {
		t.Fatalf("expected unassigned location answer to validate, got %v", err)
	}
}

func TestValidateRealtimeVoiceGeneratedResponseRequiresOperationFactsInEachChannel(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeCheckout, Operation: agentmodel.OperationCheckoutStatus,
		Subject: "loaner flashlight", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Loaner flashlight", Kind: "item", Facts: []string{"Checked out by Sam"}}},
	}
	valid := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "The loaner flashlight is checked out by Sam.",
		DisplayResponse: "The Loaner flashlight is checked out by Sam.",
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected grounded checkout response, got %v", err)
	}
	for _, invalid := range []ports.VoiceResponseGenerationResult{
		{SpokenResponse: "The loaner flashlight is available.", DisplayResponse: valid.DisplayResponse},
		{SpokenResponse: valid.SpokenResponse, DisplayResponse: "The Loaner flashlight is available."},
		{SpokenResponse: "The passport is checked out by Sam.", DisplayResponse: valid.DisplayResponse},
	} {
		if err := validateRealtimeVoiceGeneratedResponse(brief, invalid); err == nil {
			t.Fatalf("expected ungrounded checkout response to fail: %+v", invalid)
		}
	}
}

func TestValidateRealtimeVoiceGeneratedResponseRequiresNotFoundSubjectAndAbsenceInEachChannel(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeNotFound, Operation: agentmodel.OperationLocate,
		Subject: "passport", Confidence: agentmodel.ResponseConfidenceAbsent,
	}
	valid := ports.VoiceResponseGenerationResult{SpokenResponse: "I couldn't find the passport.", DisplayResponse: "I couldn't find the passport."}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected grounded not-found response, got %v", err)
	}
	invalid := ports.VoiceResponseGenerationResult{SpokenResponse: "Your passport is in the desk drawer.", DisplayResponse: valid.DisplayResponse}
	if err := validateRealtimeVoiceGeneratedResponse(brief, invalid); err == nil {
		t.Fatal("expected invented spoken location to be rejected")
	}
}

func TestValidateRealtimeVoiceGeneratedResponseAllowsBroadInventoryToUsePlacesAsContext(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeInventory, Operation: agentmodel.OperationListInventory,
		Subject: "stuff", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{
			{FactKey: "finding.0", Title: "Water bottle", Kind: "item"},
			{FactKey: "finding.1", Title: "Cordless drill", Kind: "item"},
			{FactKey: "finding.2", Title: "Empty guest room", Kind: "location"},
		},
	}
	valid := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "You have a water bottle and a cordless drill.",
		DisplayResponse: "Water bottle and Cordless drill.",
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected item-complete broad summary, got %v", err)
	}
	invalid := ports.VoiceResponseGenerationResult{SpokenResponse: "You have a water bottle.", DisplayResponse: valid.DisplayResponse}
	if err := validateRealtimeVoiceGeneratedResponse(brief, invalid); err == nil {
		t.Fatal("expected omitted inventory item to be rejected")
	}
}

func TestValidateRealtimeVoiceGeneratedResponseRejectsStateContradictions(t *testing.T) {
	t.Parallel()
	exists := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeExists, Operation: agentmodel.OperationExists,
		Subject: "drill", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Drill", Kind: "item", LifecycleState: "active", CheckoutState: "available"}},
	}
	if err := validateRealtimeVoiceGeneratedResponse(exists, ports.VoiceResponseGenerationResult{SpokenResponse: "You have a drill.", DisplayResponse: "You have a Drill."}); err != nil {
		t.Fatalf("expected positive existence answer, got %v", err)
	}
	if err := validateRealtimeVoiceGeneratedResponse(exists, ports.VoiceResponseGenerationResult{SpokenResponse: "You do not have a drill.", DisplayResponse: "You have a Drill."}); err == nil {
		t.Fatal("expected negated existence claim to fail")
	}

	checkout := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeCheckout, Operation: agentmodel.OperationCheckoutStatus,
		Subject: "flashlight", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Flashlight", Kind: "item", CheckoutState: "checked_out", Facts: []string{"Checked out by Sam"}}},
	}
	if err := validateRealtimeVoiceGeneratedResponse(checkout, ports.VoiceResponseGenerationResult{SpokenResponse: "The flashlight is not checked out by Sam.", DisplayResponse: "The Flashlight is checked out by Sam."}); err == nil {
		t.Fatal("expected negated checkout claim to fail")
	}
}

func TestValidateRealtimeVoiceGeneratedResponseAllowsPastCheckoutForAvailableAsset(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeHistory, Operation: agentmodel.OperationCheckoutHistory,
		Subject: "loaner flashlight", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{
			FactKey: "finding.0", Title: "Loaner flashlight", Kind: "item", CheckoutState: "available",
			Facts: []string{"Checked out on July 1 and returned July 3"},
		}},
	}
	result := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "The loaner flashlight was checked out on July 1 and returned July 3.",
		DisplayResponse: "The Loaner flashlight was checked out on July 1 and returned July 3.",
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, result); err != nil {
		t.Fatalf("expected historical checkout event to coexist with current availability, got %v", err)
	}
}

func TestValidateRealtimeVoiceGeneratedResponseAllowsPastArchiveForActiveAsset(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeHistory, Operation: agentmodel.OperationAssetHistory,
		Subject: "drill", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{{
			FactKey: "finding.0", Title: "Drill", Kind: "item", LifecycleState: "active",
			Facts: []string{"Archived on July 1 and restored on July 3"},
		}},
	}
	result := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "The drill was archived on July 1 and restored on July 3.",
		DisplayResponse: "The Drill was archived on July 1 and restored on July 3.",
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, result); err != nil {
		t.Fatalf("expected historical lifecycle event to coexist with current active state, got %v", err)
	}
}

func TestValidateRealtimeVoiceGeneratedResponseRequiresNaturalTruncationDisclosure(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeInventory, Operation: agentmodel.OperationListInventory,
		Subject: "stuff", Confidence: agentmodel.ResponseConfidenceStrong, Truncated: true,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Drill", Kind: "item"}},
	}
	valid := ports.VoiceResponseGenerationResult{SpokenResponse: "You have a drill and some other items.", DisplayResponse: "Drill and other items."}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected disclosed bounded summary, got %v", err)
	}
	invalid := ports.VoiceResponseGenerationResult{SpokenResponse: "You have a drill.", DisplayResponse: valid.DisplayResponse}
	if err := validateRealtimeVoiceGeneratedResponse(brief, invalid); err == nil {
		t.Fatal("expected silent truncation to fail")
	}
}

func TestValidateRealtimeVoiceGeneratedResponseAllowsGroundedLongTitleAbbreviation(t *testing.T) {
	t.Parallel()
	brief := responseComparisonBoundedSummaryBrief()
	valid := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "I found the blue storage case for letters and the green equipment bag for camping, plus other items.",
		DisplayResponse: "Blue storage case for letters; Green equipment bag for camping; and other items.",
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected distinctive long-title abbreviations to remain grounded, got %v", err)
	}
	invalid := ports.VoiceResponseGenerationResult{
		SpokenResponse:  "I found a blue case and a green bag, plus other items.",
		DisplayResponse: valid.DisplayResponse,
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, invalid); err == nil {
		t.Fatal("expected overly generic title abbreviations to fail")
	}
}

func TestValidateRealtimeVoiceGeneratedResponseRejectsInventedAbsentClarificationSuggestions(t *testing.T) {
	t.Parallel()
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: agentmodel.OperationMove,
		Subject: "passport", Confidence: agentmodel.ResponseConfidenceAbsent,
	}
	valid := ports.VoiceResponseGenerationResult{SpokenResponse: "I couldn't find your passport. Could you describe it another way?", DisplayResponse: "I couldn't find your passport. Could you describe it another way?"}
	if err := validateRealtimeVoiceGeneratedResponse(brief, valid); err != nil {
		t.Fatalf("expected bounded absent clarification, got %v", err)
	}
	invalid := ports.VoiceResponseGenerationResult{SpokenResponse: "I couldn't find your passport. Should I look in the safe?", DisplayResponse: "I couldn't find your passport. Should I look in the safe?"}
	if err := validateRealtimeVoiceGeneratedResponse(brief, invalid); err == nil {
		t.Fatal("expected invented location suggestion to be rejected")
	}
}
