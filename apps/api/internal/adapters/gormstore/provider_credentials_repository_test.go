package gormstore

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProviderCredentialRepositoryReplacesAndReadsActiveCredentialByScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	scope := providerCredentialScope("tenant-home", "profile-google")
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	first := providerCredentialRecord("credential-one", scope, "ciphertext-one", now)
	if err := store.ReplaceProviderCredential(ctx, first); err != nil {
		t.Fatalf("replace first credential: %v", err)
	}
	second := providerCredentialRecord("credential-two", scope, "ciphertext-two", now.Add(time.Minute))
	if err := store.ReplaceProviderCredential(ctx, second); err != nil {
		t.Fatalf("replace second credential: %v", err)
	}

	active, found, err := store.ActiveProviderCredential(ctx, scope)
	if err != nil {
		t.Fatalf("active credential: %v", err)
	}
	if !found || active.ID != "credential-two" || string(active.Sealed.Ciphertext) != "ciphertext-two" {
		t.Fatalf("unexpected active credential: found=%t record=%+v", found, active)
	}

	var superseded providerCredentialModel
	if err := store.db.WithContext(ctx).Where("id = ?", "credential-one").First(&superseded).Error; err != nil {
		t.Fatalf("load superseded credential: %v", err)
	}
	if superseded.SupersededAt == nil {
		t.Fatalf("expected replaced credential to be superseded")
	}
}

func TestProviderCredentialRepositoryScopesActiveCredentialReads(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	scope := providerCredentialScope("tenant-home", "profile-google")
	if err := store.ReplaceProviderCredential(ctx, providerCredentialRecord("credential-one", scope, "ciphertext-one", time.Now().UTC())); err != nil {
		t.Fatalf("replace credential: %v", err)
	}

	wrongTenant := providerCredentialScope("tenant-other", "profile-google")
	if _, found, err := store.ActiveProviderCredential(ctx, wrongTenant); err != nil || found {
		t.Fatalf("expected wrong tenant to miss, found=%t err=%v", found, err)
	}
	wrongProfile := providerCredentialScope("tenant-home", "profile-other")
	if _, found, err := store.ActiveProviderCredential(ctx, wrongProfile); err != nil || found {
		t.Fatalf("expected wrong profile to miss, found=%t err=%v", found, err)
	}
}

func TestProviderCredentialRepositorySupersedesActiveCredential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	scope := providerCredentialScope("tenant-home", "profile-google")
	if err := store.ReplaceProviderCredential(ctx, providerCredentialRecord("credential-one", scope, "ciphertext-one", time.Now().UTC())); err != nil {
		t.Fatalf("replace credential: %v", err)
	}

	if err := store.SupersedeActiveProviderCredential(ctx, scope, time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatalf("supersede credential: %v", err)
	}
	if _, found, err := store.ActiveProviderCredential(ctx, scope); err != nil || found {
		t.Fatalf("expected no active credential after supersede, found=%t err=%v", found, err)
	}
}

func TestProviderCredentialRepositoryRejectsInvalidSealedMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	scope := providerCredentialScope("tenant-home", "profile-google")

	badAlgorithm := providerCredentialRecord("credential-one", scope, "ciphertext-one", time.Now().UTC())
	badAlgorithm.Sealed.Algorithm = "plaintext"
	if err := store.ReplaceProviderCredential(ctx, badAlgorithm); !errors.Is(err, ports.ErrInvalidProviderCredential) {
		t.Fatalf("expected bad algorithm rejection, got %v", err)
	}
	badNonce := providerCredentialRecord("credential-two", scope, "ciphertext-two", time.Now().UTC())
	badNonce.Sealed.Nonce = []byte("short")
	if err := store.ReplaceProviderCredential(ctx, badNonce); !errors.Is(err, ports.ErrInvalidProviderCredential) {
		t.Fatalf("expected bad nonce rejection, got %v", err)
	}
}

func TestProviderCredentialRepositoryEnforcesOneActiveCredentialPerScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	scope := providerCredentialScope("tenant-home", "profile-google")
	now := time.Now().UTC()
	first := providerCredentialRecord("credential-one", scope, "ciphertext-one", now)
	second := providerCredentialRecord("credential-two", scope, "ciphertext-two", now)

	if err := store.db.WithContext(ctx).Create(&providerCredentialModel{
		ID:                first.ID,
		TenantID:          first.Scope.TenantID.String(),
		ProviderProfileID: first.Scope.ProviderProfileID,
		Capability:        string(first.Scope.Capability),
		ProviderKind:      string(first.Scope.ProviderKind),
		Purpose:           string(first.Scope.Purpose),
		KeyID:             first.Sealed.KeyID,
		Algorithm:         first.Sealed.Algorithm,
		Nonce:             first.Sealed.Nonce,
		Ciphertext:        first.Sealed.Ciphertext,
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("insert first active credential: %v", err)
	}
	if err := store.db.WithContext(ctx).Create(&providerCredentialModel{
		ID:                second.ID,
		TenantID:          second.Scope.TenantID.String(),
		ProviderProfileID: second.Scope.ProviderProfileID,
		Capability:        string(second.Scope.Capability),
		ProviderKind:      string(second.Scope.ProviderKind),
		Purpose:           string(second.Scope.Purpose),
		KeyID:             second.Sealed.KeyID,
		Algorithm:         second.Sealed.Algorithm,
		Nonce:             second.Sealed.Nonce,
		Ciphertext:        second.Sealed.Ciphertext,
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err == nil {
		t.Fatalf("expected duplicate active credential insert to fail")
	}
}

func TestProviderCredentialRepositoryStoresOnlySealedMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	scope := providerCredentialScope("tenant-home", "profile-google")
	rawSecret := []byte("raw-provider-secret")
	sealer, err := credentials.NewAESGCMSealer(credentials.AESGCMSealerConfig{
		KeyID:  "local-key",
		Key:    bytes.Repeat([]byte{1}, 32),
		Random: bytes.NewReader(bytes.Repeat([]byte{7}, 1024)),
	})
	if err != nil {
		t.Fatalf("new sealer: %v", err)
	}
	sealed, err := sealer.SealProviderCredential(ctx, scope, rawSecret)
	if err != nil {
		t.Fatalf("seal credential: %v", err)
	}
	record := providerCredentialRecord("credential-one", scope, "ciphertext-one", time.Now().UTC())
	record.Sealed = sealed
	if err := store.ReplaceProviderCredential(ctx, record); err != nil {
		t.Fatalf("replace credential: %v", err)
	}

	var model providerCredentialModel
	if err := store.db.WithContext(ctx).Where("id = ?", "credential-one").First(&model).Error; err != nil {
		t.Fatalf("load credential model: %v", err)
	}
	if bytes.Contains(model.Ciphertext, rawSecret) || bytes.Contains(model.Nonce, rawSecret) || model.KeyID == string(rawSecret) {
		t.Fatalf("credential row leaked raw secret: %+v", model)
	}
}

func providerCredentialScope(tenantID string, profileID string) ports.ProviderCredentialScope {
	return ports.ProviderCredentialScope{
		TenantID:          tenant.ID(tenantID),
		ProviderProfileID: profileID,
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeAPIKey,
	}
}

func providerCredentialRecord(id string, scope ports.ProviderCredentialScope, ciphertext string, now time.Time) ports.ProviderCredentialRecord {
	return ports.ProviderCredentialRecord{
		ID:    id,
		Scope: scope,
		Sealed: ports.SealedProviderCredential{
			KeyID:      "local-key",
			Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
			Nonce:      bytes.Repeat([]byte{1}, 12),
			Ciphertext: []byte(ciphertext),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
