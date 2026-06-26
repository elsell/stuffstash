package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
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
	if !found || got.ID != second.ID || got.RuntimeOptionsJSON.String() != `{"temperature":0.2}` {
		t.Fatalf("unexpected provider profile: found=%t profile=%+v", found, got)
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
