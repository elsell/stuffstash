package agentmodel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestServiceCreatesTenantProviderProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})

	profile, err := service.CreateProviderProfile(ctx, CreateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		Capability:         "language_inference",
		ProviderKind:       "gemini",
		DisplayName:        "Google Gemini",
		ModelName:          "gemini-2.5-flash-lite",
		RuntimeOptionsJSON: []byte(`{"temperature":0.1}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     "Prefer concise spoken answers.",
		Enable:             true,
	})
	if err != nil {
		t.Fatalf("create provider profile: %v", err)
	}
	if profile.TenantID.String() != "tenant-home" || profile.LifecycleState != domain.ProviderProfileEnabled {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if profile.PromptTemplate.String() != "Prefer concise spoken answers." {
		t.Fatalf("expected prompt template to be stored: %+v", profile)
	}
	if repository.saved[profile.ID.String()].DisplayName.String() != "Google Gemini" {
		t.Fatalf("expected profile to be saved: %+v", repository.saved)
	}
}

func TestServiceRejectsProviderProfileWithoutTenantConfigurePermission(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := newProviderProfileTestService(newFakeProviderProfileRepository(), denyTenantAuthorizer{})

	_, err := service.CreateProviderProfile(ctx, CreateProviderProfileInput{
		Principal:    testPrincipal(),
		Source:       audit.SourceAPI,
		RequestID:    "request-1",
		TenantID:     tenant.ID("tenant-home"),
		Capability:   "language_inference",
		ProviderKind: "gemini",
		DisplayName:  "Google Gemini",
		ModelName:    "gemini-2.5-flash-lite",
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceRejectsRawCredentialMaterialInProfileInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := newProviderProfileTestService(newFakeProviderProfileRepository(), allowTenantConfigureAuthorizer{})

	_, err := service.CreateProviderProfile(ctx, CreateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		Capability:         "language_inference",
		ProviderKind:       "gemini",
		DisplayName:        "Google Gemini",
		ModelName:          "gemini-2.5-flash-lite",
		RuntimeOptionsJSON: []byte(`{"apiKey":"secret"}`),
	})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected validation rejection, got %v", err)
	}

	for _, template := range []string{
		"Use API key value.",
		"Use secret value.",
		"Use token value.",
	} {
		_, err = service.CreateProviderProfile(ctx, CreateProviderProfileInput{
			Principal:      testPrincipal(),
			Source:         audit.SourceAPI,
			RequestID:      "request-2",
			TenantID:       tenant.ID("tenant-home"),
			Capability:     "language_inference",
			ProviderKind:   "gemini",
			DisplayName:    "Google Gemini",
			ModelName:      "gemini-2.5-flash-lite",
			PromptTemplate: template,
		})
		if !errors.Is(err, apperrors.ErrValidation) {
			t.Fatalf("expected prompt template credential rejection for %q, got %v", template, err)
		}
	}
}

func TestServiceReplacesProviderProfileCredentialWithSealedMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	updated, err := service.ReplaceProviderProfileCredential(ctx, ReplaceProviderProfileCredentialInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
		Purpose:   "api_key",
		Raw:       []byte("raw-provider-secret"),
	})
	if err != nil {
		t.Fatalf("replace provider profile credential: %v", err)
	}
	if updated.CredentialStatus != domain.CredentialStatusConfigured {
		t.Fatalf("expected configured credential status, got %+v", updated)
	}
	if string(repository.lastCredential.Sealed.Ciphertext) == "raw-provider-secret" || len(repository.lastCredential.Sealed.Ciphertext) == 0 {
		t.Fatalf("expected sealed credential material, got %+v", repository.lastCredential)
	}
	if repository.lastCredential.Scope.TenantID.String() != "tenant-home" ||
		repository.lastCredential.Scope.ProviderProfileID != "profile-one" ||
		repository.lastCredential.Scope.Capability != ports.ProviderCapabilityLanguageInference ||
		repository.lastCredential.Scope.ProviderKind != ports.ProviderKindGemini ||
		repository.lastCredential.Scope.Purpose != ports.ProviderCredentialPurposeAPIKey {
		t.Fatalf("unexpected credential scope: %+v", repository.lastCredential.Scope)
	}
}

func TestServiceReplacesProviderProfileCredentialWithServerADCMarker(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	updated, err := service.ReplaceProviderProfileCredential(ctx, ReplaceProviderProfileCredentialInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
		Purpose:   "server_adc",
	})
	if err != nil {
		t.Fatalf("replace server ADC provider profile credential: %v", err)
	}
	if updated.CredentialStatus != domain.CredentialStatusConfigured {
		t.Fatalf("expected configured credential status, got %+v", updated)
	}
	if repository.lastCredential.Scope.Purpose != ports.ProviderCredentialPurposeServerADC {
		t.Fatalf("expected server ADC credential scope, got %+v", repository.lastCredential.Scope)
	}
	if string(repository.lastCredential.Sealed.Ciphertext) != "sealed:server_adc" {
		t.Fatalf("expected sealed non-secret server ADC marker, got %+v", repository.lastCredential.Sealed)
	}
}

func TestServiceRejectsServerADCForUnsupportedProviderProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileDisabled)
	profile.ProviderKind = domain.ProviderKind("local_http")
	repository.saved[profile.ID.String()] = profile

	_, err := service.ReplaceProviderProfileCredential(ctx, ReplaceProviderProfileCredentialInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
		Purpose:   "server_adc",
	})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected validation error for unsupported server ADC purpose, got %v", err)
	}
	if repository.lastCredential.ID != "" {
		t.Fatalf("expected no credential to be persisted for unsupported server ADC purpose, got %+v", repository.lastCredential)
	}
}

func TestServiceRejectsCredentialReplacementWhenSealerUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := New(Dependencies{
		Authorizer:                allowTenantConfigureAuthorizer{},
		ProviderProfiles:          repository,
		ProviderProfileUnitOfWork: repository,
		IDs:                       fixedIDGenerator{},
		Clock:                     fixedClock{},
	})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	_, err := service.ReplaceProviderProfileCredential(ctx, ReplaceProviderProfileCredentialInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
		Purpose:   "api_key",
		Raw:       []byte("raw-provider-secret"),
	})
	if !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("expected precondition rejection, got %v", err)
	}
}

func TestServiceListsAndGetsProviderProfilesByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	first := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	second := mustProviderProfile(t, "profile-two", "tenant-home", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileEnabled)
	otherTenant := mustProviderProfile(t, "profile-three", "tenant-other", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	repository.saved[first.ID.String()] = first
	repository.saved[second.ID.String()] = second
	repository.saved[otherTenant.ID.String()] = otherTenant

	listed, err := service.ListProviderProfiles(ctx, ListProviderProfilesInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
	})
	if err != nil {
		t.Fatalf("list provider profiles: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected two tenant profiles, got %+v", listed)
	}

	got, err := service.GetProviderProfile(ctx, GetProviderProfileInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: domain.ProviderProfileID("profile-two"),
	})
	if err != nil {
		t.Fatalf("get provider profile: %v", err)
	}
	if got.ID != second.ID {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestServiceUpdatesProviderProfileConfiguration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	lastTestedAt := time.Date(2026, 6, 26, 11, 0, 0, 0, time.UTC)
	profile.LastTestedAt = &lastTestedAt
	profile.CredentialStatus = domain.CredentialStatusConfigured
	repository.saved[profile.ID.String()] = profile

	updated, err := service.UpdateProviderProfile(ctx, UpdateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		ProfileID:          profile.ID,
		DisplayName:        stringPtr("Gemini tuned"),
		EndpointURL:        stringPtr("https://generativelanguage.googleapis.com"),
		ModelName:          stringPtr("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{"temperature":0.2}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     stringPtr("Prefer concise spoken answers."),
	})
	if err != nil {
		t.Fatalf("update provider profile: %v", err)
	}
	if updated.DisplayName.String() != "Gemini tuned" || updated.RuntimeOptionsJSON.String() != `{"temperature":0.2}` || updated.PromptTemplate.String() != "Prefer concise spoken answers." {
		t.Fatalf("unexpected updated profile configuration: %+v", updated)
	}
	if updated.LastTestedAt != nil || updated.CredentialStatus != domain.CredentialStatusConfigured || updated.LifecycleState != domain.ProviderProfileEnabled {
		t.Fatalf("expected last tested reset without changing status/lifecycle, got %+v", updated)
	}
	if repository.lastAuditAction != audit.ActionProviderProfileUpdated {
		t.Fatalf("expected provider profile updated audit action, got %s", repository.lastAuditAction)
	}

	promptOnly, err := service.UpdateProviderProfile(ctx, UpdateProviderProfileInput{
		Principal:      testPrincipal(),
		Source:         audit.SourceAPI,
		RequestID:      "request-2",
		TenantID:       tenant.ID("tenant-home"),
		ProfileID:      profile.ID,
		PromptTemplate: stringPtr("Speak in one sentence."),
	})
	if err != nil {
		t.Fatalf("update provider profile prompt only: %v", err)
	}
	if promptOnly.DisplayName.String() != "Gemini tuned" || promptOnly.RuntimeOptionsJSON.String() != `{"temperature":0.2}` || promptOnly.PromptTemplate.String() != "Speak in one sentence." {
		t.Fatalf("expected partial update to preserve omitted fields, got %+v", promptOnly)
	}
}

func TestServiceRejectsProviderProfileUpdateWithoutTenantConfigurePermission(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	repository.saved[profile.ID.String()] = profile
	service := newProviderProfileTestService(repository, denyTenantAuthorizer{})

	_, err := service.UpdateProviderProfile(ctx, UpdateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		ProfileID:          profile.ID,
		DisplayName:        stringPtr("Gemini tuned"),
		RuntimeOptionsJSON: []byte(`{}`),
		CapabilityJSON:     []byte(`{}`),
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceRejectsRawCredentialMaterialInProviderProfileUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	repository.saved[profile.ID.String()] = profile
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})

	_, err := service.UpdateProviderProfile(ctx, UpdateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		ProfileID:          profile.ID,
		DisplayName:        stringPtr("Gemini tuned"),
		RuntimeOptionsJSON: []byte(`{"token":"secret"}`),
		CapabilityJSON:     []byte(`{}`),
	})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected validation rejection, got %v", err)
	}

	_, err = service.UpdateProviderProfile(ctx, UpdateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-2",
		TenantID:           tenant.ID("tenant-home"),
		ProfileID:          profile.ID,
		DisplayName:        stringPtr("Gemini tuned"),
		RuntimeOptionsJSON: []byte(`{}`),
		CapabilityJSON:     []byte(`{}`),
		PromptTemplate:     stringPtr("Use secret value."),
	})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected prompt template credential rejection, got %v", err)
	}
}

func TestServiceLifecycleCommandsRespectArchivedBoundary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	enabled, err := service.EnableProviderProfile(ctx, ProviderProfileLifecycleInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if err != nil {
		t.Fatalf("enable provider profile: %v", err)
	}
	if enabled.LifecycleState != domain.ProviderProfileEnabled {
		t.Fatalf("expected enabled profile: %+v", enabled)
	}

	archived, err := service.ArchiveProviderProfile(ctx, ProviderProfileLifecycleInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-2",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if err != nil {
		t.Fatalf("archive provider profile: %v", err)
	}
	if archived.LifecycleState != domain.ProviderProfileArchived {
		t.Fatalf("expected archived profile: %+v", archived)
	}

	if _, err := service.EnableProviderProfile(ctx, ProviderProfileLifecycleInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-3",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	}); !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("expected archived enable precondition rejection, got %v", err)
	}
}

func TestServiceTestsProviderProfileWithUnsealedCredentialAndUpdatesLastTestedAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	credentials := newFakeProviderCredentialRepository()
	tester := &fakeProviderProfileTester{}
	service := newProviderProfileTestServiceWithCredentials(repository, credentials, tester, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	profile.CredentialStatus = domain.CredentialStatusConfigured
	repository.saved[profile.ID.String()] = profile
	scope := ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "profile-one",
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeOAuthBearer,
	}
	credentials.saved[scope] = ports.ProviderCredentialRecord{
		ID:     "credential-one",
		Scope:  scope,
		Sealed: ports.SealedProviderCredential{KeyID: "test-key", Algorithm: ports.ProviderCredentialAlgorithmAES256GCM, Nonce: []byte("123456789012"), Ciphertext: []byte("sealed")},
	}
	apiKeyScope := scope
	apiKeyScope.Purpose = ports.ProviderCredentialPurposeAPIKey
	credentials.saved[apiKeyScope] = ports.ProviderCredentialRecord{
		ID:     "credential-old",
		Scope:  apiKeyScope,
		Sealed: ports.SealedProviderCredential{KeyID: "test-key", Algorithm: ports.ProviderCredentialAlgorithmAES256GCM, Nonce: []byte("123456789012"), Ciphertext: []byte("sealed-old")},
	}

	result, err := service.TestProviderProfile(ctx, TestProviderProfileInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if err != nil {
		t.Fatalf("test provider profile: %v", err)
	}
	if result.Status != ports.ProviderProfileTestStatusSucceeded || result.ProfileID != "profile-one" || result.Message == "" {
		t.Fatalf("unexpected provider test result: %+v", result)
	}
	if tester.lastInput.Profile.ID != profile.ID || string(tester.lastInput.Credential) != "raw:profile-one" || tester.lastInput.CredentialPurpose != ports.ProviderCredentialPurposeAPIKey {
		t.Fatalf("tester did not receive expected provider input: %+v", tester.lastInput)
	}
	updated := repository.saved[profile.ID.String()]
	if updated.LastTestedAt == nil || !updated.LastTestedAt.Equal(fixedClock{}.Now()) {
		t.Fatalf("expected last tested timestamp update, got %+v", updated.LastTestedAt)
	}
	if repository.lastAuditAction != audit.ActionProviderProfileTested {
		t.Fatalf("expected provider profile tested audit action, got %s", repository.lastAuditAction)
	}
}

func TestServiceProviderCredentialPurposesPreferServerADCForGeminiTextToSpeech(t *testing.T) {
	t.Parallel()

	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileDisabled)
	got := providerCredentialPurposes(profile)
	if len(got) != 2 || got[0] != ports.ProviderCredentialPurposeServerADC || got[1] != ports.ProviderCredentialPurposeOAuthBearer {
		t.Fatalf("unexpected Gemini text-to-speech credential order: %+v", got)
	}
}

func TestServiceProviderCredentialPurposesPreferAPIKeyThenServerADCForGeminiLanguage(t *testing.T) {
	t.Parallel()

	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	got := providerCredentialPurposes(profile)
	if len(got) != 3 ||
		got[0] != ports.ProviderCredentialPurposeAPIKey ||
		got[1] != ports.ProviderCredentialPurposeServerADC ||
		got[2] != ports.ProviderCredentialPurposeOAuthBearer {
		t.Fatalf("unexpected Gemini language credential order: %+v", got)
	}
}

func TestServiceReturnsSafeFailedProviderProfileTestResultAndWritesAudit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	credentials := newFakeProviderCredentialRepository()
	tester := &fakeProviderProfileTester{
		result: ports.ProviderProfileTestResult{
			Status:  ports.ProviderProfileTestStatusFailed,
			Message: "provider account secret-detail quota stack trace",
		},
	}
	service := newProviderProfileTestServiceWithCredentials(repository, credentials, tester, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	profile.CredentialStatus = domain.CredentialStatusConfigured
	repository.saved[profile.ID.String()] = profile
	scope := ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "profile-one",
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeOAuthBearer,
	}
	credentials.saved[scope] = ports.ProviderCredentialRecord{
		ID:     "credential-one",
		Scope:  scope,
		Sealed: ports.SealedProviderCredential{KeyID: "test-key", Algorithm: ports.ProviderCredentialAlgorithmAES256GCM, Nonce: []byte("123456789012"), Ciphertext: []byte("sealed")},
	}

	result, err := service.TestProviderProfile(ctx, TestProviderProfileInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if err != nil {
		t.Fatalf("test provider profile: %v", err)
	}
	if result.Status != ports.ProviderProfileTestStatusFailed || strings.Contains(result.Message, "secret-detail") || result.TestedAt.IsZero() {
		t.Fatalf("expected safe failed provider test result, got %+v", result)
	}
	updated := repository.saved[profile.ID.String()]
	if updated.LastTestedAt != nil {
		t.Fatalf("failed provider test should not update last tested timestamp, got %+v", updated.LastTestedAt)
	}
	if repository.lastAuditAction != audit.ActionProviderProfileTested {
		t.Fatalf("expected provider profile tested audit action for failed test, got %s", repository.lastAuditAction)
	}
}

func TestServiceRejectsProviderProfileTestWithoutConfiguredCredential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestServiceWithCredentials(repository, newFakeProviderCredentialRepository(), &fakeProviderProfileTester{}, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	_, err := service.TestProviderProfile(ctx, TestProviderProfileInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("expected precondition rejection, got %v", err)
	}
}

func TestServiceRejectsProviderProfileTestWithInvalidStoredCredential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	credentials := newFakeProviderCredentialRepository()
	service := New(Dependencies{
		Authorizer:                allowTenantConfigureAuthorizer{},
		ProviderProfiles:          repository,
		ProviderProfileUnitOfWork: repository,
		ProviderCredentialVault:   fakeCredentialVault{repository: credentials, sealer: fakeCredentialSealer{}, activeErr: ports.ErrInvalidProviderInput},
		ProviderProfileTester:     &fakeProviderProfileTester{},
		IDs:                       fixedIDGenerator{},
		Clock:                     fixedClock{},
	})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	repository.saved[profile.ID.String()] = profile
	scope := ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "profile-one",
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeAPIKey,
	}
	credentials.saved[scope] = ports.ProviderCredentialRecord{ID: "credential-one", Scope: scope}

	_, err := service.TestProviderProfile(ctx, TestProviderProfileInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("expected safe precondition for invalid stored credential, got %v", err)
	}
}
