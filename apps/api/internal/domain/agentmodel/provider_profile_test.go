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

func newTestProviderProfile(t *testing.T, lifecycle ProviderProfileLifecycleState) ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := NewProviderProfile(ProviderProfileInput{
		ID:                 ProviderProfileID("profile-google"),
		TenantID:           TenantID("tenant-home"),
		Capability:         ProviderCapabilityLanguageInference,
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
