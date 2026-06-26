package voice

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ProviderProfileTester struct {
	factory ProviderProfileProviderFactory
}

func NewProviderProfileTester(factory ProviderProfileProviderFactory) ProviderProfileTester {
	return ProviderProfileTester{factory: factory}
}

func (t ProviderProfileTester) TestProviderProfile(ctx context.Context, input ports.ProviderProfileTestInput) (ports.ProviderProfileTestResult, error) {
	if t.factory == nil || len(input.Credential) == 0 || input.TestedAt.IsZero() {
		return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
	}
	config := ProviderProfileProviderConfig{
		Profile:           input.Profile,
		CredentialPurpose: input.CredentialPurpose,
		Credential:        append([]byte{}, input.Credential...),
	}
	switch input.Profile.Capability {
	case agentmodel.ProviderCapabilitySpeechToText:
		provider, err := t.factory.SpeechToTextProvider(ctx, config)
		if err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
		if provider == nil {
			return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
		}
		if err := probeSpeechToTextProvider(ctx, provider); err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
	case agentmodel.ProviderCapabilityLanguageInference:
		provider, err := t.factory.LanguageInferenceProvider(ctx, config)
		if err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
		if provider == nil {
			return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
		}
		if err := probeLanguageInferenceProvider(ctx, provider); err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
	case agentmodel.ProviderCapabilityTextToSpeech:
		provider, err := t.factory.TextToSpeechProvider(ctx, config)
		if err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
		if provider == nil {
			return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
		}
		if err := probeTextToSpeechProvider(ctx, provider); err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
	default:
		return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
	}
	return ports.ProviderProfileTestResult{
		ProfileID:    input.Profile.ID.String(),
		Capability:   input.Profile.Capability.String(),
		ProviderKind: input.Profile.ProviderKind.String(),
		Status:       ports.ProviderProfileTestStatusSucceeded,
		Message:      safeProviderProfileTestMessage(input.Profile),
		TestedAt:     input.TestedAt,
	}, nil
}

func probeSpeechToTextProvider(ctx context.Context, provider ports.SpeechToTextProvider) error {
	if probe, ok := provider.(ports.SpeechToTextProviderProbe); ok {
		return probe.ProbeSpeechToText(ctx)
	}
	return ports.ErrInvalidProviderInput
}

func probeLanguageInferenceProvider(ctx context.Context, provider ports.LanguageInferenceProvider) error {
	if probe, ok := provider.(ports.LanguageInferenceProviderProbe); ok {
		return probe.ProbeLanguageInference(ctx)
	}
	turn, err := provider.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript: "Provider diagnostic. Return a final answer that says Provider profile test succeeded.",
		FinalOnly:  true,
	})
	if err != nil {
		return err
	}
	if turn.Final == nil || strings.TrimSpace(turn.Final.SpokenResponse) == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func probeTextToSpeechProvider(ctx context.Context, provider ports.TextToSpeechProvider) error {
	if probe, ok := provider.(ports.TextToSpeechProviderProbe); ok {
		return probe.ProbeTextToSpeech(ctx)
	}
	result, err := provider.Synthesize(ctx, ports.TextToSpeechInput{Text: "Stuff Stash provider test."})
	if err != nil {
		return err
	}
	if strings.TrimSpace(result.MimeType) == "" || len(result.Chunks) == 0 {
		return ports.ErrInvalidProviderInput
	}
	for _, chunk := range result.Chunks {
		if len(chunk) > 0 {
			return nil
		}
	}
	return ports.ErrInvalidProviderInput
}

func safeProviderProfileTestMessage(profile agentmodel.ProviderProfile) string {
	name := strings.TrimSpace(profile.DisplayName.String())
	if name == "" {
		name = "Provider profile"
	}
	return name + " completed a safe provider diagnostic."
}
