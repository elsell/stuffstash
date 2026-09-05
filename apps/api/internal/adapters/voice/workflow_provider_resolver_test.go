package voice

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestWorkflowProviderResolverRequiresExactUsableScopedProfile(t *testing.T) {
	for _, test := range []struct {
		name       string
		tenant     string
		id         string
		capability agentmodel.ProviderCapability
		lifecycle  agentmodel.ProviderProfileLifecycleState
		valid      bool
	}{
		{"selected", "tenant-home", "selected", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, true},
		{"cross tenant", "tenant-other", "selected", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, false},
		{"unknown", "tenant-home", "missing", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, false},
		{"disabled", "tenant-home", "selected", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled, false},
		{"wrong capability", "tenant-home", "selected", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			profile := providerResolverProfile(t, "selected", test.capability, test.lifecycle, agentmodel.CredentialStatusConfigured)
			vault := newProviderResolverCredentialVault(profile)
			factory := &providerResolverFactory{}
			resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: []agentmodel.ProviderProfile{profile}}, nil, vault, factory)
			binding, err := resolver.ResolveWorkflowLanguageProvider(context.Background(), ports.WorkflowLanguageProviderResolutionInput{TenantID: tenant.ID(test.tenant), ProfileID: test.id})
			if test.valid {
				if err != nil || binding.ProfileID != "selected" || binding.Provider == nil || binding.PromptTemplate != "Prefer concise spoken answers." {
					t.Fatalf("explicit binding: %+v %v", binding, err)
				}
			} else {
				if !errors.Is(err, ports.ErrInvalidProviderInput) {
					t.Fatalf("invalid selection accepted: %v", err)
				}
				if len(vault.scopes) != 0 || len(factory.configs) != 0 {
					t.Fatal("invalid selection reached credentials or factory")
				}
			}
		})
	}
}

func TestWorkflowCanResolveSpeechWithoutAnUnusedDefaultLanguageModel(t *testing.T) {
	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	factory := &providerResolverFactory{}
	resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: profiles}, nil, newProviderResolverCredentialVault(profiles...), factory)
	set, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: "tenant-home", SkipDefaultLanguage: true})
	if err != nil || set.SpeechToText == nil || set.TextToSpeech == nil || set.LanguageInference != nil {
		t.Fatalf("unused default LM still required: %+v %v", set, err)
	}
	if len(factory.configs) != 2 {
		t.Fatalf("unexpected provider construction: %d", len(factory.configs))
	}
	if _, err = resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: "tenant-home"}); !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatal("default flow must still require LM")
	}
}
