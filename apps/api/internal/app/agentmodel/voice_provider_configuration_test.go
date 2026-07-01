package agentmodel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestServiceBuildsVoiceProviderConfigurationDiagnostics(t *testing.T) {
	t.Parallel()

	repository := newFakeProviderProfileRepository()
	stt := providerProfileForVoiceConfig(t, "stt-ready", domain.ProviderCapabilitySpeechToText, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, testedAt())
	duplicate := providerProfileForVoiceConfig(t, "stt-duplicate", domain.ProviderCapabilitySpeechToText, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, testedAt())
	language := providerProfileForVoiceConfig(t, "language-missing-credential", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled, domain.CredentialStatusMissing, nil)
	tts := providerProfileForVoiceConfigWithRuntimeOptions(t, "tts-untested", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, nil, `{"credentialType":"server_adc"}`)
	for _, profile := range []domain.ProviderProfile{stt, duplicate, language, tts} {
		repository.saved[profile.ID.String()] = profile
	}
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})

	view, err := service.GetVoiceProviderConfiguration(context.Background(), VoiceProviderConfigurationInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		TenantID:  tenant.ID("tenant-home"),
	})
	if err != nil {
		t.Fatalf("get voice provider configuration: %v", err)
	}
	if view.Readiness != VoiceProviderReadinessNeedsAttention {
		t.Fatalf("expected needs attention readiness, got %+v", view)
	}
	if view.Slots[0].Capability != domain.ProviderCapabilitySpeechToText.String() || view.Slots[0].SelectionSource != VoiceProviderSelectionImplicit || view.Slots[0].Readiness != VoiceProviderSlotDuplicateCandidates {
		t.Fatalf("unexpected speech slot: %+v", view.Slots[0])
	}
	if len(view.Slots[0].DuplicateProfiles) != 2 {
		t.Fatalf("expected duplicate speech candidates, got %+v", view.Slots[0].DuplicateProfiles)
	}
	if view.Slots[1].Readiness != VoiceProviderSlotCredentialMissing {
		t.Fatalf("expected language credential warning, got %+v", view.Slots[1])
	}
	if view.Slots[2].Readiness != VoiceProviderSlotUntested {
		t.Fatalf("expected tts untested warning, got %+v", view.Slots[2])
	}
	if view.Slots[2].SelectedProfile == nil || view.Slots[2].SelectedProfile.CredentialPurpose != string(ports.ProviderCredentialPurposeServerADC) {
		t.Fatalf("expected tts selected profile to expose server ADC credential purpose, got %+v", view.Slots[2].SelectedProfile)
	}
}

func TestServiceSavesExplicitVoiceProviderConfiguration(t *testing.T) {
	t.Parallel()

	repository := newFakeProviderProfileRepository()
	stt := providerProfileForVoiceConfig(t, "stt-ready", domain.ProviderCapabilitySpeechToText, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, testedAt())
	language := providerProfileForVoiceConfig(t, "language-ready", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, testedAt())
	tts := providerProfileForVoiceConfig(t, "tts-ready", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, testedAt())
	for _, profile := range []domain.ProviderProfile{stt, language, tts} {
		repository.saved[profile.ID.String()] = profile
	}
	observer := &fakeObserver{}
	service := newProviderProfileTestServiceWithObserver(repository, newFakeProviderCredentialRepository(), &fakeProviderProfileTester{}, allowTenantConfigureAuthorizer{}, observer)

	view, err := service.UpdateVoiceProviderConfiguration(context.Background(), UpdateVoiceProviderConfigurationInput{
		Principal:                  testPrincipal(),
		Source:                     audit.SourceAPI,
		RequestID:                  "request-voice-config",
		TenantID:                   tenant.ID("tenant-home"),
		SpeechToTextProfileID:      "stt-ready",
		LanguageInferenceProfileID: "language-ready",
		TextToSpeechProfileID:      "tts-ready",
	})
	if err != nil {
		t.Fatalf("update voice provider configuration: %v", err)
	}
	if view.Readiness != VoiceProviderReadinessReady {
		t.Fatalf("expected ready configuration, got %+v", view)
	}
	for _, slot := range view.Slots {
		if slot.SelectionSource != VoiceProviderSelectionExplicit || slot.Readiness != VoiceProviderSlotReady {
			t.Fatalf("expected explicit ready slot, got %+v", slot)
		}
	}
	if repository.voiceConfig.SpeechToTextProfileID != "stt-ready" || repository.lastAuditAction != audit.ActionVoiceProviderConfigurationUpdated {
		t.Fatalf("expected persisted voice configuration and audit action, got config=%+v action=%s", repository.voiceConfig, repository.lastAuditAction)
	}
	if len(observer.events) != 1 || observer.events[0].Name != ports.EventVoiceProviderConfigurationUpdated {
		t.Fatalf("expected voice provider configuration observability event, got %+v", observer.events)
	}
}

func TestServiceRejectsWrongCapabilityVoiceProviderConfiguration(t *testing.T) {
	t.Parallel()

	repository := newFakeProviderProfileRepository()
	repository.saved["language-ready"] = providerProfileForVoiceConfig(t, "language-ready", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled, domain.CredentialStatusConfigured, testedAt())
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})

	_, err := service.UpdateVoiceProviderConfiguration(context.Background(), UpdateVoiceProviderConfigurationInput{
		Principal:             testPrincipal(),
		Source:                audit.SourceAPI,
		TenantID:              tenant.ID("tenant-home"),
		SpeechToTextProfileID: "language-ready",
	})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestServiceReportsArchivedSelectedVoiceProviderProfile(t *testing.T) {
	t.Parallel()

	repository := newFakeProviderProfileRepository()
	archived := providerProfileForVoiceConfig(t, "stt-archived", domain.ProviderCapabilitySpeechToText, domain.ProviderProfileArchived, domain.CredentialStatusConfigured, testedAt())
	repository.saved[archived.ID.String()] = archived
	repository.voiceConfig = ports.VoiceProviderConfigurationRecord{
		TenantID:              tenant.ID("tenant-home"),
		SpeechToTextProfileID: "stt-archived",
	}
	repository.voiceConfigSet = true
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})

	view, err := service.GetVoiceProviderConfiguration(context.Background(), VoiceProviderConfigurationInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		TenantID:  tenant.ID("tenant-home"),
	})
	if err != nil {
		t.Fatalf("get voice provider configuration: %v", err)
	}
	if view.Slots[0].Readiness != VoiceProviderSlotArchived {
		t.Fatalf("expected archived diagnostic, got %+v", view.Slots[0])
	}
}

func providerProfileForVoiceConfig(t *testing.T, id string, capability domain.ProviderCapability, lifecycle domain.ProviderProfileLifecycleState, credential domain.CredentialStatus, tested *time.Time) domain.ProviderProfile {
	t.Helper()
	return providerProfileForVoiceConfigWithRuntimeOptions(t, id, capability, lifecycle, credential, tested, `{}`)
}

func providerProfileForVoiceConfigWithRuntimeOptions(t *testing.T, id string, capability domain.ProviderCapability, lifecycle domain.ProviderProfileLifecycleState, credential domain.CredentialStatus, tested *time.Time, runtimeOptions string) domain.ProviderProfile {
	t.Helper()
	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := domain.NewProviderProfile(domain.ProviderProfileInput{
		ID:                 domain.ProviderProfileID(id),
		TenantID:           domain.TenantID("tenant-home"),
		Capability:         capability,
		ProviderKind:       domain.ProviderKindGemini,
		DisplayName:        domain.DisplayName(id),
		ModelName:          domain.ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(runtimeOptions),
		CapabilityJSON:     []byte(`{}`),
		CredentialStatus:   credential,
		LifecycleState:     lifecycle,
		LastTestedAt:       tested,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("invalid provider profile fixture")
	}
	return profile
}

func testedAt() *time.Time {
	value := time.Date(2026, 6, 26, 11, 0, 0, 0, time.UTC)
	return &value
}
