package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestVoiceProviderConfigurationRepositorySavesReadsAndScopesByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")

	record := ports.VoiceProviderConfigurationRecord{
		TenantID:                   tenant.ID("tenant-home"),
		SpeechToTextProfileID:      "profile-stt",
		LanguageInferenceProfileID: "profile-lm",
		TextToSpeechProfileID:      "profile-tts",
		CreatedAt:                  time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
		UpdatedAt:                  time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
	}
	if err := store.SaveVoiceProviderConfiguration(ctx, record, auditRecord(t, "audit-voice-config-one", tenant.ID("tenant-home"), "", audit.ActionVoiceProviderConfigurationUpdated)); err != nil {
		t.Fatalf("save voice provider configuration: %v", err)
	}

	got, found, err := store.VoiceProviderConfiguration(ctx, tenant.ID("tenant-home"))
	if err != nil {
		t.Fatalf("get voice provider configuration: %v", err)
	}
	if !found || got.SpeechToTextProfileID != "profile-stt" || got.LanguageInferenceProfileID != "profile-lm" || got.TextToSpeechProfileID != "profile-tts" {
		t.Fatalf("unexpected voice provider configuration: found=%t record=%+v", found, got)
	}
	if _, found, err := store.VoiceProviderConfiguration(ctx, tenant.ID("tenant-other")); err != nil || found {
		t.Fatalf("expected other tenant miss, found=%t err=%v", found, err)
	}

	updated := record
	updated.SpeechToTextProfileID = "profile-stt-two"
	updated.UpdatedAt = record.UpdatedAt.Add(time.Minute)
	if err := store.SaveVoiceProviderConfiguration(ctx, updated, auditRecord(t, "audit-voice-config-two", tenant.ID("tenant-home"), "", audit.ActionVoiceProviderConfigurationUpdated)); err != nil {
		t.Fatalf("update voice provider configuration: %v", err)
	}
	got, found, err = store.VoiceProviderConfiguration(ctx, tenant.ID("tenant-home"))
	if err != nil {
		t.Fatalf("get updated voice provider configuration: %v", err)
	}
	if !found || got.SpeechToTextProfileID != "profile-stt-two" || !got.CreatedAt.Equal(record.CreatedAt) {
		t.Fatalf("unexpected updated voice provider configuration: found=%t record=%+v", found, got)
	}
}

func TestVoiceProviderConfigurationRepositoryRollsBackOnAuditFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")

	auditID := "audit-voice-config"
	if err := store.SaveAuditRecord(ctx, auditRecord(t, auditID, tenant.ID("tenant-home"), "", audit.ActionVoiceProviderConfigurationUpdated)); err != nil {
		t.Fatalf("save existing audit record: %v", err)
	}
	record := ports.VoiceProviderConfigurationRecord{
		TenantID:                   tenant.ID("tenant-home"),
		SpeechToTextProfileID:      "profile-stt",
		LanguageInferenceProfileID: "profile-lm",
		TextToSpeechProfileID:      "profile-tts",
		CreatedAt:                  time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
		UpdatedAt:                  time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
	}
	if err := store.SaveVoiceProviderConfiguration(ctx, record, auditRecord(t, auditID, tenant.ID("tenant-home"), "", audit.ActionVoiceProviderConfigurationUpdated)); err == nil {
		t.Fatalf("expected duplicate audit insert to fail")
	}
	if _, found, err := store.VoiceProviderConfiguration(ctx, tenant.ID("tenant-home")); err != nil || found {
		t.Fatalf("expected voice provider configuration rollback, found=%t err=%v", found, err)
	}
}
