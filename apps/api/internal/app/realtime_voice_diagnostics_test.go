package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceDiagnosticRedactionCoversHeaderBearerTokens(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		"Authorization: bearer abc/def==",
		"token: bearer abc/def==",
		"authorization=bearer abc/def==",
	} {
		redacted := safeRealtimeVoiceDiagnosticText(input, 500)
		if strings.Contains(redacted, "abc/def") || strings.Contains(strings.ToLower(redacted), "bearer ") || !strings.Contains(redacted, "[redacted") {
			t.Fatalf("expected bearer material to be redacted, input %q became %q", input, redacted)
		}
	}
}

func TestRealtimeVoiceTypedInvestigationDiagnosticsRespectSessionSetting(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{false, true} {
		enabled := enabled
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			t.Parallel()
			initial, final := realtimeVoiceTypedLocateTurns("diagnostic-subject", "Diagnostic subject")
			language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
				{Investigation: &initial},
				{Investigation: &final},
			}}
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated diagnostic request"}
			resolver.providers.LanguageInference = language
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("diagnostic-subject", "Diagnostic subject", asset.KindItem, ""), "audit-diagnostic-subject")
			input := defaultRealtimeVoiceSessionInput()
			input.DeveloperDiagnostics = enabled
			session, err := application.StartRealtimeVoiceSession(context.Background(), input)
			if err != nil {
				t.Fatalf("start session: %v", err)
			}
			events := []RealtimeVoiceEvent{}
			if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
				events = append(events, event)
				return nil
			}); err != nil {
				t.Fatalf("run voice entrypoint: %v", err)
			}
			diagnostics := []RealtimeVoiceEvent{}
			for _, event := range events {
				if event.Type == RealtimeVoiceEventAgentDiagnostic {
					diagnostics = append(diagnostics, event)
				}
			}
			if !enabled && len(diagnostics) != 0 {
				t.Fatalf("diagnostics-disabled session leaked diagnostics: %+v", diagnostics)
			}
			if enabled {
				if len(diagnostics) != 2 {
					t.Fatalf("expected typed-turn diagnostics, got %+v", events)
				}
				for _, diagnostic := range diagnostics {
					for _, forbidden := range []string{"generated diagnostic request", "Diagnostic subject", "diagnostic-subject", "search_assets", "subjectMention"} {
						if strings.Contains(diagnostic.Detail, forbidden) {
							t.Fatalf("diagnostic leaked %q: %+v", forbidden, diagnostic)
						}
					}
					if !strings.Contains(diagnostic.Detail, `"operation":"locate"`) || !strings.Contains(diagnostic.Message, "Language investigation") {
						t.Fatalf("expected safe typed diagnostic metadata: %+v", diagnostic)
					}
				}
			}
		})
	}
}

func TestRealtimeVoiceTypedContinuationFailureRetainsSafeReadDiagnostics(t *testing.T) {
	t.Parallel()
	initial, _ := realtimeVoiceTypedLocateTurns("failure-subject", "Failure subject")
	language := &scriptedRealtimeLanguageInference{
		turns: []ports.LanguageInferenceTurn{{Investigation: &initial}},
		errs:  []error{nil, safeRealtimeVoiceDiagnosticFailure{safe: "provider_http_status_429"}},
	}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated continuation-failure request"}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("failure-subject", "Failure subject", asset.KindItem, ""), "audit-failure-subject")
	input := defaultRealtimeVoiceSessionInput()
	input.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), input)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err == nil || RealtimeVoiceSafeErrorCode(err) != realtimeVoiceFailureLanguageInference {
		t.Fatalf("expected language inference stage failure, got %v", err)
	}
	diagnostic := findRealtimeVoiceDiagnosticEvent(t, events, "Language provider failed")
	for _, required := range []string{`"phase": "evidence_assessment"`, `"evidenceRound": 1`, `"maxEvidenceRounds": 2`, `"previousRequestCount": 1`, `"observationCount": 1`, `"readEvidenceCount": 1`, `"toolResultCount": 1`, RealtimeVoiceToolSearchAuthorizedAssets, "provider_http_status_429"} {
		if !strings.Contains(diagnostic.Detail, required) {
			t.Fatalf("expected safe continuation diagnostic to contain %q, got %s", required, diagnostic.Detail)
		}
	}
	if strings.Contains(diagnostic.Detail, "provider.invalid") || strings.Contains(diagnostic.Detail, "should-not-leak") || strings.Contains(diagnostic.Detail, "finalOnly") || strings.Contains(diagnostic.Detail, "previousTurns") {
		t.Fatalf("provider internals leaked through diagnostic: %s", diagnostic.Detail)
	}
}

func TestRealtimeVoiceTypedResponsePreservesTextToSpeechBoundaries(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name          string
		provider      ports.TextToSpeechProvider
		wantFailure   bool
		wantSafeError string
		wantChunks    int
	}{
		{name: "provider failure", provider: failingResolvedTextToSpeech{err: safeRealtimeVoiceDiagnosticFailure{safe: "provider_timeout"}}, wantFailure: true, wantSafeError: "provider_timeout"},
		{name: "malformed output", provider: malformedResolvedTextToSpeech{}, wantFailure: true, wantSafeError: "invalid_provider_output"},
		{name: "empty chunks compacted", provider: resolvedTextToSpeechWithChunks{chunks: [][]byte{[]byte("speech"), nil}}, wantChunks: 1},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			initial, final := realtimeVoiceTypedLocateTurns("speech-subject", "Speech subject")
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated speech-boundary request"}
			resolver.providers.LanguageInference = &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}
			resolver.providers.TextToSpeech = testCase.provider
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("speech-subject", "Speech subject", asset.KindItem, ""), "audit-speech-subject")
			input := defaultRealtimeVoiceSessionInput()
			input.DeveloperDiagnostics = true
			session, err := application.StartRealtimeVoiceSession(context.Background(), input)
			if err != nil {
				t.Fatalf("start session: %v", err)
			}
			events := []RealtimeVoiceEvent{}
			err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
				events = append(events, event)
				return nil
			})
			if testCase.wantFailure {
				if err == nil || RealtimeVoiceSafeErrorCode(err) != realtimeVoiceFailureTextToSpeech {
					t.Fatalf("expected text-to-speech stage failure, got %v", err)
				}
				diagnostic := findRealtimeVoiceDiagnosticEvent(t, events, "Text-to-speech provider failed")
				if !strings.Contains(diagnostic.Detail, testCase.wantSafeError) || !strings.Contains(diagnostic.Detail, `"toolResultCount": 1`) {
					t.Fatalf("expected safe TTS diagnostic, got %s", diagnostic.Detail)
				}
				return
			}
			if err != nil {
				t.Fatalf("run voice entrypoint: %v", err)
			}
			audioChunks := []RealtimeVoiceEvent{}
			for _, event := range events {
				if event.Type == RealtimeVoiceEventTextToSpeechAudioChunk {
					audioChunks = append(audioChunks, event)
				}
			}
			if len(audioChunks) != testCase.wantChunks || !audioChunks[len(audioChunks)-1].FinalChunk {
				t.Fatalf("expected compacted final speech chunks, got %+v", audioChunks)
			}
		})
	}
}

func realtimeVoiceTypedLocateTurns(candidateID, title string) (agentmodel.InvestigationStep, agentmodel.InvestigationStep) {
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: title}
	initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: title, SearchProbes: []string{title},
	}}}
	final := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{candidateID},
	}}}
	return initial, final
}

func findRealtimeVoiceDiagnosticEvent(t *testing.T, events []RealtimeVoiceEvent, message string) RealtimeVoiceEvent {
	t.Helper()
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == message {
			return event
		}
	}
	t.Fatalf("expected diagnostic %q, got %+v", message, events)
	return RealtimeVoiceEvent{}
}

type safeRealtimeVoiceDiagnosticFailure struct{ safe string }

func (e safeRealtimeVoiceDiagnosticFailure) Error() string {
	return "provider failure with raw endpoint https://provider.invalid bearer should-not-leak"
}

func (e safeRealtimeVoiceDiagnosticFailure) SafeRealtimeVoiceDiagnostic() string { return e.safe }

type failingResolvedTextToSpeech struct{ err error }

func (f failingResolvedTextToSpeech) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{}, f.err
}

type malformedResolvedTextToSpeech struct{}

func (malformedResolvedTextToSpeech) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{nil}}, nil
}

type resolvedTextToSpeechWithChunks struct{ chunks [][]byte }

func (r resolvedTextToSpeechWithChunks) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: r.chunks}, nil
}
