package voice

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type DevFakeSpeechToText struct{}

func (DevFakeSpeechToText) Transcribe(_ context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	return ports.SpeechToTextResult{Transcript: "Where are my tools?"}, nil
}

type DevFakeLanguageInference struct{}

func (DevFakeLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if input.Investigation == nil || input.Investigation.Validate() != nil {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if input.Investigation.Phase == agentmodel.InvestigationPhaseInitial {
		return ports.LanguageInferenceTurn{Investigation: &agentmodel.InvestigationStep{
			Decision: agentmodel.InvestigationDecisionSearch,
			Intent:   agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "tools"},
			SearchRequests: []agentmodel.SearchRequest{{
				ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets,
				Mention: "tools", SearchProbes: []string{"tools"}, LifecycleScope: agentmodel.LifecycleScopeActive,
			}},
			Rationale: "Gather visible candidates.",
		}}, nil
	}
	resolution := agentmodel.Resolution{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionAbsent, Evidence: "No visible candidate matched."}
	if len(input.Investigation.Observations) > 0 {
		resolution.Status = agentmodel.ResolutionPlausible
		resolution.CandidateIDs = []string{input.Investigation.Observations[0].CandidateID}
		resolution.Evidence = "A visible candidate matched."
	}
	return ports.LanguageInferenceTurn{Investigation: &agentmodel.InvestigationStep{
		Decision:    agentmodel.InvestigationDecisionFinish,
		Intent:      *input.Investigation.CanonicalIntent,
		Resolutions: []agentmodel.Resolution{resolution},
		Rationale:   "Resolve from authorized evidence.",
	}}, nil
}

func (DevFakeLanguageInference) ProbeLanguageInference(context.Context) error { return nil }

func (DevFakeLanguageInference) GenerateResponse(_ context.Context, input ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	if input.Brief.Validate() != nil {
		return ports.VoiceResponseGenerationResult{}, ports.ErrInvalidProviderInput
	}
	text := "I couldn't find that in this inventory."
	if len(input.Brief.Findings) > 0 {
		finding := input.Brief.Findings[0]
		location := finding.Title
		if len(finding.ContainmentPath) > 0 {
			location = finding.ContainmentPath[len(finding.ContainmentPath)-1]
		}
		prefix := "I found it"
		if input.Brief.Confidence == agentmodel.ResponseConfidencePlausible {
			prefix = "I think it's"
		}
		text = prefix + " in " + location + "."
	}
	return ports.VoiceResponseGenerationResult{SpokenResponse: text, DisplayResponse: text}, nil
}

type DevFakeTextToSpeech struct{}

func (DevFakeTextToSpeech) Synthesize(_ context.Context, input ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	if input.Text == "" {
		return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
	}
	return ports.TextToSpeechResult{
		MimeType: "audio/mpeg",
		Chunks:   [][]byte{[]byte(input.Text)},
	}, nil
}
