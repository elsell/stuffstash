package voice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProviderProfileTesterInitializesProviderAdapter(t *testing.T) {
	t.Parallel()

	profile := providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured)
	testedAt := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	tester := NewProviderProfileTester(&providerResolverFactory{})

	result, err := tester.TestProviderProfile(context.Background(), ports.ProviderProfileTestInput{
		Profile:           profile,
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("token"),
		TestedAt:          testedAt,
	})
	if err != nil {
		t.Fatalf("test provider profile: %v", err)
	}
	if result.ProfileID != "lm-profile" || result.Status != ports.ProviderProfileTestStatusSucceeded || !result.TestedAt.Equal(testedAt) {
		t.Fatalf("unexpected provider profile test result: %+v", result)
	}
}

func TestProviderProfileTesterRunsCapabilityProbeWhenAvailable(t *testing.T) {
	t.Parallel()

	testedAt := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	factory := &probeProviderProfileFactory{}
	tester := NewProviderProfileTester(factory)
	cases := []struct {
		name       string
		capability agentmodel.ProviderCapability
		probed     *bool
	}{
		{name: "speech to text", capability: agentmodel.ProviderCapabilitySpeechToText, probed: &factory.stt.probed},
		{name: "language inference", capability: agentmodel.ProviderCapabilityLanguageInference, probed: &factory.language.probed},
		{name: "text to speech", capability: agentmodel.ProviderCapabilityTextToSpeech, probed: &factory.tts.probed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := providerResolverProfile(t, tc.name, tc.capability, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured)
			result, err := tester.TestProviderProfile(context.Background(), ports.ProviderProfileTestInput{
				Profile:           profile,
				CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
				Credential:        []byte("token"),
				TestedAt:          testedAt,
			})
			if err != nil {
				t.Fatalf("test provider profile: %v", err)
			}
			if !*tc.probed || result.Status != ports.ProviderProfileTestStatusSucceeded {
				t.Fatalf("expected successful probed result, probed=%v result=%+v", *tc.probed, result)
			}
		})
	}
}

func TestProviderProfileTesterReturnsProbeFailure(t *testing.T) {
	t.Parallel()

	profile := providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured)
	factory := &probeProviderProfileFactory{}
	factory.language.err = ports.ErrInvalidProviderInput
	tester := NewProviderProfileTester(factory)

	_, err := tester.TestProviderProfile(context.Background(), ports.ProviderProfileTestInput{
		Profile:           profile,
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("token"),
		TestedAt:          time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected probe failure, got %v", err)
	}
}

func TestProviderProfileTesterRejectsSpeechToTextAdapterWithoutProbe(t *testing.T) {
	t.Parallel()

	profile := providerResolverProfile(t, "stt-profile", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured)
	tester := NewProviderProfileTester(nonProbeSpeechToTextFactory{})

	_, err := tester.TestProviderProfile(context.Background(), ports.ProviderProfileTestInput{
		Profile:           profile,
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("token"),
		TestedAt:          time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
}

func TestProviderProfileTesterRejectsNilProviderAdapter(t *testing.T) {
	t.Parallel()

	profile := providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured)
	tester := NewProviderProfileTester(nilProviderProfileFactory{})

	_, err := tester.TestProviderProfile(context.Background(), ports.ProviderProfileTestInput{
		Profile:           profile,
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("token"),
		TestedAt:          time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
}

type nilProviderProfileFactory struct{}

func (nilProviderProfileFactory) SpeechToTextProvider(context.Context, ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	return nil, nil
}

func (nilProviderProfileFactory) LanguageInferenceProvider(context.Context, ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error) {
	return nil, nil
}

func (nilProviderProfileFactory) TextToSpeechProvider(context.Context, ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	return nil, nil
}

type nonProbeSpeechToTextFactory struct{}

func (nonProbeSpeechToTextFactory) SpeechToTextProvider(context.Context, ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	return providerResolverSpeechToText{}, nil
}

func (nonProbeSpeechToTextFactory) LanguageInferenceProvider(context.Context, ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error) {
	return providerResolverLanguageInference{}, nil
}

func (nonProbeSpeechToTextFactory) TextToSpeechProvider(context.Context, ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	return providerResolverTextToSpeech{}, nil
}

type probeProviderProfileFactory struct {
	stt      probeSpeechToTextDouble
	language probeLanguageInferenceDouble
	tts      probeTextToSpeechDouble
}

func (f *probeProviderProfileFactory) SpeechToTextProvider(context.Context, ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	return &f.stt, nil
}

func (f *probeProviderProfileFactory) LanguageInferenceProvider(context.Context, ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error) {
	return &f.language, nil
}

func (f *probeProviderProfileFactory) TextToSpeechProvider(context.Context, ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	return &f.tts, nil
}

type probeSpeechToTextDouble struct {
	probed bool
	err    error
}

func (p *probeSpeechToTextDouble) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	return ports.SpeechToTextResult{Transcript: "diagnostic"}, nil
}

func (p *probeSpeechToTextDouble) ProbeSpeechToText(context.Context) error {
	p.probed = true
	return p.err
}

type probeLanguageInferenceDouble struct {
	probed bool
	err    error
}

func (p *probeLanguageInferenceDouble) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: "ready", DisplayResponse: "ready"}}, nil
}

func (p *probeLanguageInferenceDouble) ProbeLanguageInference(context.Context) error {
	p.probed = true
	return p.err
}

type probeTextToSpeechDouble struct {
	probed bool
	err    error
}

func (p *probeTextToSpeechDouble) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{[]byte("audio")}}, nil
}

func (p *probeTextToSpeechDouble) ProbeTextToSpeech(context.Context) error {
	p.probed = true
	return p.err
}
