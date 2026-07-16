package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProviderProfileRepositorySavesListsAndGetsByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	first := providerProfile(t, "profile-one", "tenant-home", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled)
	second := providerProfile(t, "profile-two", "tenant-home", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled)
	other := providerProfile(t, "profile-three", "tenant-other", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled)

	if err := store.SaveProviderProfile(ctx, first, auditRecord(t, "audit-profile-one", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save first provider profile: %v", err)
	}
	if err := store.SaveProviderProfile(ctx, second, auditRecord(t, "audit-profile-two", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save second provider profile: %v", err)
	}
	if err := store.SaveProviderProfile(ctx, other, auditRecord(t, "audit-profile-three", tenant.ID("tenant-other"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save other provider profile: %v", err)
	}

	listed, err := store.ListProviderProfiles(ctx, tenant.ID("tenant-home"))
	if err != nil {
		t.Fatalf("list provider profiles: %v", err)
	}
	if len(listed) != 2 || listed[0].ID != first.ID || listed[1].ID != second.ID {
		t.Fatalf("unexpected tenant profiles: %+v", listed)
	}

	got, found, err := store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), second.ID)
	if err != nil {
		t.Fatalf("get provider profile: %v", err)
	}
	if !found || got.ID != second.ID || got.RuntimeOptionsJSON.String() != `{"temperature":0.2}` || got.PromptTemplate.String() != "" {
		t.Fatalf("unexpected provider profile: found=%t profile=%+v", found, got)
	}
	got, found, err = store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), first.ID)
	if err != nil {
		t.Fatalf("get language provider profile: %v", err)
	}
	if !found || got.PromptTemplate.String() != "Answer in one short sentence." {
		t.Fatalf("expected language profile prompt template, found=%t profile=%+v", found, got)
	}
	if _, found, err := store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), other.ID); err != nil || found {
		t.Fatalf("expected cross-tenant profile miss, found=%t err=%v", found, err)
	}
}

func TestProviderProfileRepositoryUpdatesLifecycleAndPreservesAuditAtomicity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	profile := providerProfile(t, "profile-one", "tenant-home", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled)
	auditID := "audit-profile-one"
	if err := store.SaveAuditRecord(ctx, auditRecord(t, auditID, tenant.ID("tenant-home"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save existing audit record: %v", err)
	}
	if err := store.SaveProviderProfile(ctx, profile, auditRecord(t, "audit-profile-two", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save provider profile: %v", err)
	}

	enabled, ok := profile.Enable(time.Now().UTC().Add(time.Minute))
	if !ok {
		t.Fatalf("enable profile")
	}
	if err := store.UpdateProviderProfile(ctx, enabled, auditRecord(t, auditID, tenant.ID("tenant-home"), "", audit.ActionProviderProfileEnabled)); err == nil {
		t.Fatalf("expected duplicate audit insert to fail")
	}
	got, found, err := store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), profile.ID)
	if err != nil {
		t.Fatalf("get provider profile after failed update: %v", err)
	}
	if !found || got.LifecycleState != agentmodel.ProviderProfileDisabled {
		t.Fatalf("expected profile update to roll back after audit failure: found=%t profile=%+v", found, got)
	}

	if err := store.UpdateProviderProfile(ctx, enabled, auditRecord(t, "audit-profile-three", tenant.ID("tenant-home"), "", audit.ActionProviderProfileEnabled)); err != nil {
		t.Fatalf("update provider profile: %v", err)
	}
	got, found, err = store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), profile.ID)
	if err != nil {
		t.Fatalf("get provider profile after update: %v", err)
	}
	if !found || got.LifecycleState != agentmodel.ProviderProfileEnabled {
		t.Fatalf("expected enabled provider profile: found=%t profile=%+v", found, got)
	}

	lastTestedAt := enabled.UpdatedAt.Add(time.Minute)
	enabled.LastTestedAt = &lastTestedAt
	configured, ok := enabled.UpdateConfiguration(agentmodel.ProviderProfileConfigurationUpdate{
		DisplayName:        agentmodel.DisplayName("Google Gemini tuned"),
		EndpointURL:        agentmodel.EndpointURL("https://generativelanguage.googleapis.com"),
		ModelName:          agentmodel.ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{"temperature":0.3}`),
		CapabilityJSON:     []byte(`{"toolCalls":true,"json":true}`),
		PromptTemplate:     "Answer briefly.",
		UpdatedAt:          lastTestedAt.Add(time.Minute),
	})
	if !ok {
		t.Fatalf("update provider profile configuration")
	}
	if err := store.UpdateProviderProfile(ctx, configured, auditRecord(t, "audit-profile-four", tenant.ID("tenant-home"), "", audit.ActionProviderProfileUpdated)); err != nil {
		t.Fatalf("update provider profile configuration: %v", err)
	}
	got, found, err = store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), profile.ID)
	if err != nil {
		t.Fatalf("get provider profile after configuration update: %v", err)
	}
	if !found || got.DisplayName.String() != "Google Gemini tuned" || got.RuntimeOptionsJSON.String() != `{"temperature":0.3}` || got.CapabilityJSON.String() != `{"toolCalls":true,"json":true}` || got.PromptTemplate.String() != "Answer briefly." || got.LastTestedAt != nil {
		t.Fatalf("expected updated provider profile configuration: found=%t profile=%+v", found, got)
	}
}

func TestProviderProfileRepositoryReplacesCredentialAndUpdatesProfileStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	profile := providerProfile(t, "profile-one", "tenant-home", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled)
	if err := store.SaveProviderProfile(ctx, profile, auditRecord(t, "audit-profile-one", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save provider profile: %v", err)
	}

	configured, ok := profile.WithCredentialConfigured(time.Now().UTC().Add(time.Minute))
	if !ok {
		t.Fatalf("configure credential status")
	}
	first := providerProfileCredential("credential-one", configured, "ciphertext-one", configured.UpdatedAt)
	if err := store.ReplaceProviderProfileCredential(ctx, configured, first, auditRecord(t, "audit-credential-one", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCredentialReplaced)); err != nil {
		t.Fatalf("replace provider credential: %v", err)
	}
	second := providerProfileCredential("credential-two", configured, "ciphertext-two", configured.UpdatedAt.Add(time.Minute))
	if err := store.ReplaceProviderProfileCredential(ctx, configured, second, auditRecord(t, "audit-credential-two", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCredentialReplaced)); err != nil {
		t.Fatalf("replace provider credential again: %v", err)
	}

	got, found, err := store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), profile.ID)
	if err != nil {
		t.Fatalf("get provider profile: %v", err)
	}
	if !found || got.CredentialStatus != agentmodel.CredentialStatusConfigured {
		t.Fatalf("expected configured provider profile: found=%t profile=%+v", found, got)
	}
	active, found, err := store.ActiveProviderCredential(ctx, second.Scope)
	if err != nil {
		t.Fatalf("active provider credential: %v", err)
	}
	if !found || active.ID != "credential-two" || string(active.Sealed.Ciphertext) != "ciphertext-two" {
		t.Fatalf("unexpected active credential: found=%t credential=%+v", found, active)
	}
	var superseded providerCredentialModel
	if err := store.db.WithContext(ctx).Where(&providerCredentialModel{ID: "credential-one"}).First(&superseded).Error; err != nil {
		t.Fatalf("load superseded credential: %v", err)
	}
	if superseded.SupersededAt == nil {
		t.Fatalf("expected prior credential to be superseded")
	}
}

func TestProviderProfileRepositoryRejectsCredentialScopeMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	profile := providerProfile(t, "profile-one", "tenant-home", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileDisabled)
	if err := store.SaveProviderProfile(ctx, profile, auditRecord(t, "audit-profile-one", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save provider profile: %v", err)
	}

	configured, ok := profile.WithCredentialConfigured(time.Now().UTC().Add(time.Minute))
	if !ok {
		t.Fatalf("configure credential status")
	}
	mismatched := providerProfileCredential("credential-one", configured, "ciphertext-one", configured.UpdatedAt)
	mismatched.Scope.TenantID = tenant.ID("tenant-other")
	if err := store.ReplaceProviderProfileCredential(ctx, configured, mismatched, auditRecord(t, "audit-credential-one", tenant.ID("tenant-home"), "", audit.ActionProviderProfileCredentialReplaced)); err != ports.ErrInvalidProviderCredential {
		t.Fatalf("expected invalid provider credential, got %v", err)
	}

	got, found, err := store.ProviderProfileByID(ctx, tenant.ID("tenant-home"), profile.ID)
	if err != nil {
		t.Fatalf("get provider profile: %v", err)
	}
	if !found || got.CredentialStatus != agentmodel.CredentialStatusMissing {
		t.Fatalf("expected credential status to remain missing: found=%t profile=%+v", found, got)
	}
	if exists, err := store.ActiveProviderCredentialsExist(ctx); err != nil || exists {
		t.Fatalf("expected no active credentials after mismatch: exists=%t err=%v", exists, err)
	}
}

func providerProfile(t *testing.T, id string, tenantID string, capability agentmodel.ProviderCapability, lifecycle agentmodel.ProviderProfileLifecycleState) agentmodel.ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := agentmodel.NewProviderProfile(agentmodel.ProviderProfileInput{
		ID:                 agentmodel.ProviderProfileID(id),
		TenantID:           agentmodel.TenantID(tenantID),
		Capability:         capability,
		ProviderKind:       agentmodel.ProviderKindGemini,
		DisplayName:        agentmodel.DisplayName("Google Gemini"),
		EndpointURL:        agentmodel.EndpointURL("https://generativelanguage.googleapis.com"),
		ModelName:          agentmodel.ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{"temperature":0.2}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     providerProfilePromptTemplate(capability),
		CredentialStatus:   agentmodel.CredentialStatusMissing,
		LifecycleState:     lifecycle,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("expected valid provider profile")
	}
	return profile
}

func providerProfilePromptTemplate(capability agentmodel.ProviderCapability) string {
	if capability == agentmodel.ProviderCapabilityLanguageInference {
		return "Answer in one short sentence."
	}
	return ""
}

func providerProfileCredential(id string, profile agentmodel.ProviderProfile, ciphertext string, now time.Time) ports.ProviderCredentialRecord {
	return ports.ProviderCredentialRecord{
		ID: id,
		Scope: ports.ProviderCredentialScope{
			TenantID:          tenant.ID(profile.TenantID.String()),
			ProviderProfileID: profile.ID.String(),
			Capability:        ports.ProviderCapability(profile.Capability.String()),
			ProviderKind:      ports.ProviderKind(profile.ProviderKind.String()),
			Purpose:           ports.ProviderCredentialPurposeAPIKey,
		},
		Sealed: ports.SealedProviderCredential{
			KeyID:      "local-key",
			Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
			Nonce:      []byte("123456789012"),
			Ciphertext: []byte(ciphertext),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
