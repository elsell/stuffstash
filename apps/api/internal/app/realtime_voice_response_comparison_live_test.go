package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLiveResponseRealizationBeatsDeterministicBaseline(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("set STUFF_STASH_GOOGLE_LIVE_TESTS=1 to run the live response comparison")
	}
	projectID := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		t.Skip("set STUFF_STASH_GOOGLE_CLOUD_PROJECT to run the live response comparison")
	}
	location := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_LOCATION"))
	if location == "" {
		location = "us-central1"
	}
	model := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_GEMINI_MODEL"))
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	provider := voice.NewGoogleGeminiLanguageInference(voice.GoogleGeminiConfig{
		ProjectID: projectID, Location: location, Model: model, QuotaProject: projectID,
		TokenSource: liveGoogleTokenSource(t, ctx), HTTPTimeout: 120 * time.Second,
	})

	type trial struct {
		name  string
		brief agentmodel.GroundedVoiceResponseBrief
	}
	trials := []trial{
		{name: "category container location", brief: responseComparisonBrief(agentmodel.ResponseAnswerModeLocate, agentmodel.OperationLocate, "tools", agentmodel.ResponseConfidencePlausible,
			agentmodel.ResponseFinding{FactKey: "finding.0", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "Toolbox"}})},
		{name: "known item location", brief: responseComparisonBrief(agentmodel.ResponseAnswerModeLocate, agentmodel.OperationLocate, "water bottle", agentmodel.ResponseConfidenceStrong,
			agentmodel.ResponseFinding{FactKey: "finding.0", Title: "Water bottle", Kind: "item", ContainmentPath: []string{"Office", "Water bottle"}})},
		{name: "location contents", brief: responseComparisonBrief(agentmodel.ResponseAnswerModeContents, agentmodel.OperationListContents, "Toolbox", agentmodel.ResponseConfidenceStrong,
			agentmodel.ResponseFinding{FactKey: "finding.0", Title: "Cordless drill", Kind: "item"}, agentmodel.ResponseFinding{FactKey: "finding.1", Title: "Flashlight", Kind: "item"})},
		{name: "inventory summary", brief: responseComparisonBrief(agentmodel.ResponseAnswerModeInventory, agentmodel.OperationListInventory, "items", agentmodel.ResponseConfidenceStrong,
			agentmodel.ResponseFinding{FactKey: "finding.0", Title: "Garden shears", Kind: "item"}, agentmodel.ResponseFinding{FactKey: "finding.1", Title: "Water bottle", Kind: "item"})},
		{name: "larger inventory summary", brief: responseComparisonBrief(agentmodel.ResponseAnswerModeInventory, agentmodel.OperationListInventory, "items", agentmodel.ResponseConfidenceStrong,
			agentmodel.ResponseFinding{FactKey: "finding.0", Title: "Office", Kind: "location"}, agentmodel.ResponseFinding{FactKey: "finding.1", Title: "Living room", Kind: "location"},
			agentmodel.ResponseFinding{FactKey: "finding.2", Title: "Garage", Kind: "location"}, agentmodel.ResponseFinding{FactKey: "finding.3", Title: "Toolbox", Kind: "container"},
			agentmodel.ResponseFinding{FactKey: "finding.4", Title: "Water bottle", Kind: "item"}, agentmodel.ResponseFinding{FactKey: "finding.5", Title: "Cordless drill", Kind: "item"},
			agentmodel.ResponseFinding{FactKey: "finding.6", Title: "Garden shears", Kind: "item"}, agentmodel.ResponseFinding{FactKey: "finding.7", Title: "Loaner flashlight", Kind: "item"})},
		{name: "bounded maximum summary", brief: responseComparisonBoundedSummaryBrief()},
		{name: "approximate remembered title", brief: responseComparisonBrief(agentmodel.ResponseAnswerModeLocate, agentmodel.OperationLocate, "Sarah winter coat", agentmodel.ResponseConfidencePlausible,
			agentmodel.ResponseFinding{FactKey: "finding.0", Title: "Sarah Winter Clothes and Shoes", Kind: "container", ContainmentPath: []string{"Basement", "Storage room", "Sarah Winter Clothes and Shoes"}})},
		{name: "ambiguous title", brief: agentmodel.GroundedVoiceResponseBrief{
			Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: agentmodel.OperationLocate,
			Subject: "drill", Confidence: agentmodel.ResponseConfidenceAmbiguous,
			Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Blue drill", Kind: "item"}, {FactKey: "finding.1", Title: "Red drill", Kind: "item"}},
		}},
		{name: "no match", brief: agentmodel.GroundedVoiceResponseBrief{
			Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeNotFound, Operation: agentmodel.OperationLocate,
			Subject: "passport", Confidence: agentmodel.ResponseConfidenceAbsent,
		}},
		{name: "missing existing source", brief: agentmodel.GroundedVoiceResponseBrief{
			Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: agentmodel.OperationMove,
			Subject: "passport", Confidence: agentmodel.ResponseConfidenceAbsent,
		}},
	}

	generatedValid := 0
	baselineValid := 0
	generatedMechanical := 0
	baselineMechanical := 0
	latencies := []time.Duration{}
	for _, trial := range trials {
		for repetition := 0; repetition < 3; repetition++ {
			started := time.Now()
			generated, err := provider.GenerateResponse(ctx, ports.VoiceResponseGenerationInput{Brief: trial.brief})
			latency := time.Since(started)
			if err != nil {
				t.Fatalf("%s repetition %d generation failed: %v", trial.name, repetition+1, err)
			}
			latencies = append(latencies, latency)
			generatedValidResult := responseComparisonIndependentSemanticValid(trial.brief, generated) && validateRealtimeVoiceGeneratedResponse(trial.brief, generated) == nil
			if generatedValidResult {
				generatedValid++
			}
			if realtimeVoiceComparisonMechanical(generated.SpokenResponse) {
				generatedMechanical++
			}

			baseline := deterministicVoiceResponseBaseline(trial.brief)
			baselineValidResult := responseComparisonIndependentSemanticValid(trial.brief, baseline)
			if baselineValidResult {
				baselineValid++
			}
			if realtimeVoiceComparisonMechanical(baseline.SpokenResponse) {
				baselineMechanical++
			}
			t.Logf("comparison scenario=%q repetition=%d generated=%q baseline=%q generated_valid=%t baseline_valid=%t latency=%s",
				trial.name, repetition+1, generated.SpokenResponse, baseline.SpokenResponse, generatedValidResult, baselineValidResult, latency)
		}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := latencies[len(latencies)/2]
	p95 := latencies[(len(latencies)-1)*95/100]
	total := len(trials) * 3
	t.Logf("comparison summary model=%s trials=%d generated_semantic=%d/%d baseline_semantic=%d/%d generated_mechanical=%d baseline_mechanical=%d latency_p50=%s latency_p95=%s",
		model, total, generatedValid, total, baselineValid, total, generatedMechanical, baselineMechanical, p50, p95)
	if generatedValid != total {
		t.Fatalf("generated realization must pass every semantic gate, got %d/%d", generatedValid, total)
	}
	if generatedMechanical != 0 {
		t.Fatalf("generated realization used implementation-oriented language in %d trials", generatedMechanical)
	}
	if generatedValid <= baselineValid || baselineMechanical == 0 {
		t.Fatalf("expected generated realization to improve on baseline: generated=%d baseline=%d baseline_mechanical=%d", generatedValid, baselineValid, baselineMechanical)
	}
}

func responseComparisonBrief(mode agentmodel.ResponseAnswerMode, operation agentmodel.Operation, subject string, confidence agentmodel.ResponseConfidence, findings ...agentmodel.ResponseFinding) agentmodel.GroundedVoiceResponseBrief {
	return agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: mode, Operation: operation,
		Subject: subject, Confidence: confidence, Findings: findings,
	}
}

func responseComparisonBoundedSummaryBrief() agentmodel.GroundedVoiceResponseBrief {
	return agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeInventory, Operation: agentmodel.OperationListInventory,
		Subject: "stored items", Confidence: agentmodel.ResponseConfidenceStrong, Truncated: true,
		Findings: []agentmodel.ResponseFinding{
			{FactKey: "finding.0", Title: "Blue archival storage case for family photographs and handwritten letters from summer trips", Kind: "item"},
			{FactKey: "finding.1", Title: "Green equipment bag for camping lanterns cooking utensils and water purification supplies", Kind: "item"},
		},
	}
}

func responseComparisonIndependentSemanticValid(brief agentmodel.GroundedVoiceResponseBrief, result ports.VoiceResponseGenerationResult) bool {
	channels := []string{strings.TrimSpace(result.SpokenResponse), strings.TrimSpace(result.DisplayResponse)}
	for index, channel := range channels {
		limit := 500
		if index == 1 {
			limit = 1000
		}
		if channel == "" || len(channel) > limit || realtimeVoiceComparisonMechanical(channel) {
			return false
		}
		lower := strings.ToLower(channel)
		if (brief.Kind == agentmodel.ResponseBriefKindClarification) != strings.Contains(channel, "?") {
			return false
		}
		if brief.Confidence == agentmodel.ResponseConfidencePlausible && !responseComparisonContainsAny(lower, "probably", "likely", "might", "may ", "i think", "seems", "could ") {
			return false
		}
		if brief.Truncated && !responseComparisonContainsAny(lower, "other", "more", "additional", "including", "among") {
			return false
		}
		required := brief.Findings
		if brief.Mode == agentmodel.ResponseAnswerModeInventory {
			items := make([]agentmodel.ResponseFinding, 0, len(required))
			for _, finding := range required {
				if finding.Kind == "item" {
					items = append(items, finding)
				}
			}
			if len(items) > 0 {
				required = items
			}
		}
		for _, finding := range required {
			if !responseComparisonContainsFinding(lower, finding.Title) {
				return false
			}
		}
		switch brief.Mode {
		case agentmodel.ResponseAnswerModeLocate:
			if !responseComparisonContainsAny(" "+lower+" ", " in ", " at ", " inside ", " under ", " on ", " within ") {
				return false
			}
			for _, finding := range brief.Findings {
				if finding.Kind == "item" && len(finding.ContainmentPath) > 1 && !strings.Contains(lower, strings.ToLower(finding.ContainmentPath[len(finding.ContainmentPath)-2])) {
					return false
				}
			}
		case agentmodel.ResponseAnswerModeNotFound:
			if !strings.Contains(lower, strings.ToLower(brief.Subject)) || !responseComparisonContainsAny(lower, "can't find", "couldn't find", "not found", "no match") {
				return false
			}
		case agentmodel.ResponseAnswerModeClarify:
			if len(brief.Findings) == 0 && !strings.Contains(lower, strings.ToLower(brief.Subject)) {
				return false
			}
		}
	}
	return true
}

func responseComparisonContainsAny(value string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func responseComparisonContainsFinding(lower string, title string) bool {
	if strings.Contains(lower, strings.ToLower(title)) {
		return true
	}
	if len(title) <= 64 {
		return false
	}
	words := strings.FieldsFunc(strings.ToLower(title), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	stop := map[string]bool{"the": true, "and": true, "for": true, "from": true, "with": true, "this": true, "that": true}
	anchors := make([]string, 0, len(words))
	for _, word := range words {
		if !stop[word] && len([]rune(word)) >= 3 {
			anchors = append(anchors, word)
		}
	}
	if len(anchors) == 0 || !strings.Contains(lower, anchors[0]) {
		return false
	}
	matches := 0
	for _, anchor := range anchors {
		if strings.Contains(lower, anchor) {
			matches++
		}
	}
	return matches >= min(3, len(anchors))
}

// deterministicVoiceResponseBaseline preserves the removed renderer only as a
// test comparator. It is not wired into production and cannot answer a session.
func deterministicVoiceResponseBaseline(brief agentmodel.GroundedVoiceResponseBrief) ports.VoiceResponseGenerationResult {
	titles := make([]string, 0, len(brief.Findings))
	for _, finding := range brief.Findings {
		titles = append(titles, finding.Title)
	}
	message := "I couldn't find a visible match in this inventory."
	switch brief.Mode {
	case agentmodel.ResponseAnswerModeLocate:
		if len(brief.Findings) == 1 && brief.Confidence == agentmodel.ResponseConfidenceStrong {
			message = brief.Findings[0].Title + ". Its recorded path is " + strings.Join(brief.Findings[0].ContainmentPath, " / ") + "."
		} else {
			message = fmt.Sprintf("I found %d visible matches: %s.", len(titles), strings.Join(titles, ", "))
		}
	case agentmodel.ResponseAnswerModeInventory, agentmodel.ResponseAnswerModeContents:
		message = fmt.Sprintf("I found %d visible matches: %s.", len(titles), strings.Join(titles, ", "))
	case agentmodel.ResponseAnswerModeClarify:
		message = "I found multiple plausible matches: " + strings.Join(titles, "; ") + ". Which one did you mean?"
	}
	return ports.VoiceResponseGenerationResult{SpokenResponse: message, DisplayResponse: message}
}

func realtimeVoiceComparisonMechanical(text string) bool {
	normalized := strings.ToLower(text)
	for _, term := range []string{"visible match", "recorded path", "candidate", "resolution", " / "} {
		if strings.Contains(normalized, term) {
			return true
		}
	}
	return false
}
