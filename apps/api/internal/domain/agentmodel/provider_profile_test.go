package agentmodel

import (
	"testing"
	"time"
)

func TestNewProviderProfileBuildsTypedTenantScopedProfile(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := NewProviderProfile(ProviderProfileInput{
		ID:                 ProviderProfileID("profile-google"),
		TenantID:           TenantID("tenant-home"),
		Capability:         ProviderCapabilityLanguageInference,
		ProviderKind:       ProviderKindGemini,
		DisplayName:        DisplayName("Google Gemini"),
		EndpointURL:        EndpointURL("https://generativelanguage.googleapis.com"),
		ModelName:          ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{"temperature":0.1}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     "Prefer concise spoken answers.",
		CredentialStatus:   CredentialStatusMissing,
		LifecycleState:     ProviderProfileDisabled,
		CreatedAt:          created,
		UpdatedAt:          created,
	})

	if !ok {
		t.Fatalf("expected provider profile to validate")
	}
	if profile.ID.String() != "profile-google" || profile.TenantID.String() != "tenant-home" {
		t.Fatalf("unexpected profile identity: %+v", profile)
	}
	if profile.RuntimeOptionsJSON.String() != `{"temperature":0.1}` || profile.CapabilityJSON.String() != `{"toolCalls":true}` {
		t.Fatalf("expected JSON metadata to be preserved: %+v", profile)
	}
	if profile.PromptTemplate.String() != "Prefer concise spoken answers." {
		t.Fatalf("expected prompt template to be preserved: %+v", profile)
	}
}

func TestNewProviderProfileRejectsInvalidLifecycleAndMetadata(t *testing.T) {
	t.Parallel()

	valid := ProviderProfileInput{
		ID:                 ProviderProfileID("profile-google"),
		TenantID:           TenantID("tenant-home"),
		Capability:         ProviderCapabilityLanguageInference,
		ProviderKind:       ProviderKindGemini,
		DisplayName:        DisplayName("Google Gemini"),
		ModelName:          ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{}`),
		CapabilityJSON:     []byte(`{}`),
		CredentialStatus:   CredentialStatusMissing,
		LifecycleState:     ProviderProfileDisabled,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	invalidLifecycle := valid
	invalidLifecycle.LifecycleState = ProviderProfileLifecycleState("active")
	if _, ok := NewProviderProfile(invalidLifecycle); ok {
		t.Fatalf("expected invalid lifecycle rejection")
	}

	invalidMetadata := valid
	invalidMetadata.RuntimeOptionsJSON = []byte(`[]`)
	if _, ok := NewProviderProfile(invalidMetadata); ok {
		t.Fatalf("expected non-object runtime options rejection")
	}

	invalidPromptCapability := valid
	invalidPromptCapability.Capability = ProviderCapabilityTextToSpeech
	invalidPromptCapability.PromptTemplate = "Prefer concise spoken answers."
	if _, ok := NewProviderProfile(invalidPromptCapability); ok {
		t.Fatalf("expected prompt template on non-language profile rejection")
	}

	invalidPromptLength := valid
	invalidPromptLength.PromptTemplate = string(make([]byte, 8193))
	if _, ok := NewProviderProfile(invalidPromptLength); ok {
		t.Fatalf("expected oversized prompt template rejection")
	}
}

func TestProviderProfileLifecycleTransitions(t *testing.T) {
	t.Parallel()

	profile := newTestProviderProfile(t, ProviderProfileDisabled)
	enabled, ok := profile.Enable(time.Now().UTC())
	if !ok || enabled.LifecycleState != ProviderProfileEnabled {
		t.Fatalf("expected disabled profile to enable: %+v ok=%t", enabled, ok)
	}
	disabled, ok := enabled.Disable(time.Now().UTC())
	if !ok || disabled.LifecycleState != ProviderProfileDisabled {
		t.Fatalf("expected enabled profile to disable: %+v ok=%t", disabled, ok)
	}
	archived, ok := disabled.Archive(time.Now().UTC())
	if !ok || archived.LifecycleState != ProviderProfileArchived {
		t.Fatalf("expected disabled profile to archive: %+v ok=%t", archived, ok)
	}
	if _, ok := archived.Enable(time.Now().UTC()); ok {
		t.Fatalf("expected archived profile not to re-enable")
	}
}

func TestProviderProfileUpdatesConfigurationAndClearsLastTestedAt(t *testing.T) {
	t.Parallel()

	profile := newTestProviderProfile(t, ProviderProfileEnabled)
	testedAt := time.Date(2026, 6, 26, 11, 0, 0, 0, time.UTC)
	profile.LastTestedAt = &testedAt
	updatedAt := testedAt.Add(time.Hour)

	updated, ok := profile.UpdateConfiguration(ProviderProfileConfigurationUpdate{
		DisplayName:        DisplayName("Gemini tuned"),
		EndpointURL:        EndpointURL("https://generativelanguage.googleapis.com"),
		ModelName:          ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{"temperature":0.2}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     "Prefer concise spoken answers.",
		UpdatedAt:          updatedAt,
	})

	if !ok {
		t.Fatalf("expected configuration update to validate")
	}
	if updated.ID != profile.ID || updated.Capability != profile.Capability || updated.ProviderKind != profile.ProviderKind || updated.CredentialStatus != profile.CredentialStatus || updated.LifecycleState != profile.LifecycleState {
		t.Fatalf("configuration update changed stable fields: before=%+v after=%+v", profile, updated)
	}
	if updated.DisplayName.String() != "Gemini tuned" || updated.RuntimeOptionsJSON.String() != `{"temperature":0.2}` || updated.PromptTemplate.String() != "Prefer concise spoken answers." {
		t.Fatalf("configuration update did not preserve editable fields: %+v", updated)
	}
	if updated.LastTestedAt != nil || !updated.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected last tested reset and updated timestamp, got lastTested=%+v updatedAt=%s", updated.LastTestedAt, updated.UpdatedAt)
	}
}

func TestProviderProfileRejectsInvalidConfigurationUpdates(t *testing.T) {
	t.Parallel()

	profile := newTestProviderProfile(t, ProviderProfileEnabled)
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)

	invalidMetadata := ProviderProfileConfigurationUpdate{
		DisplayName:        DisplayName("Gemini tuned"),
		RuntimeOptionsJSON: []byte(`[]`),
		CapabilityJSON:     []byte(`{}`),
		UpdatedAt:          now,
	}
	if _, ok := profile.UpdateConfiguration(invalidMetadata); ok {
		t.Fatalf("expected invalid runtime options rejection")
	}

	archived, ok := profile.Archive(now)
	if !ok {
		t.Fatalf("archive profile")
	}
	valid := ProviderProfileConfigurationUpdate{
		DisplayName:        DisplayName("Gemini tuned"),
		RuntimeOptionsJSON: []byte(`{}`),
		CapabilityJSON:     []byte(`{}`),
		UpdatedAt:          now.Add(time.Hour),
	}
	if _, ok := archived.UpdateConfiguration(valid); ok {
		t.Fatalf("expected archived profile update rejection")
	}

	tts := newTestProviderProfileWithCapability(t, ProviderCapabilityTextToSpeech, ProviderProfileEnabled)
	valid.PromptTemplate = "Prefer concise spoken answers."
	if _, ok := tts.UpdateConfiguration(valid); ok {
		t.Fatalf("expected prompt template on non-language profile rejection")
	}
}

func newTestProviderProfile(t *testing.T, lifecycle ProviderProfileLifecycleState) ProviderProfile {
	t.Helper()

	return newTestProviderProfileWithCapability(t, ProviderCapabilityLanguageInference, lifecycle)
}

func newTestProviderProfileWithCapability(t *testing.T, capability ProviderCapability, lifecycle ProviderProfileLifecycleState) ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := NewProviderProfile(ProviderProfileInput{
		ID:                 ProviderProfileID("profile-google"),
		TenantID:           TenantID("tenant-home"),
		Capability:         capability,
		ProviderKind:       ProviderKindGemini,
		DisplayName:        DisplayName("Google Gemini"),
		ModelName:          ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{}`),
		CapabilityJSON:     []byte(`{}`),
		CredentialStatus:   CredentialStatusMissing,
		LifecycleState:     lifecycle,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("expected test provider profile to validate")
	}
	return profile
}
