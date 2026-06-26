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
	case agentmodel.ProviderCapabilityLanguageInference:
		provider, err := t.factory.LanguageInferenceProvider(ctx, config)
		if err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
		if provider == nil {
			return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
		}
	case agentmodel.ProviderCapabilityTextToSpeech:
		provider, err := t.factory.TextToSpeechProvider(ctx, config)
		if err != nil {
			return ports.ProviderProfileTestResult{}, err
		}
		if provider == nil {
			return ports.ProviderProfileTestResult{}, ports.ErrInvalidProviderInput
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

func safeProviderProfileTestMessage(profile agentmodel.ProviderProfile) string {
	name := strings.TrimSpace(profile.DisplayName.String())
	if name == "" {
		name = "Provider profile"
	}
	return name + " is configured well enough to initialize its provider adapter."
}
