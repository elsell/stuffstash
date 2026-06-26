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
