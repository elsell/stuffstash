package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProviderCredentialRepositoryReplacesAndSupersedesActiveCredential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewStore()
	scope := memoryProviderCredentialScope()
	first := memoryProviderCredential("credential-one", scope, "ciphertext-one", time.Now().UTC())
	if err := store.ReplaceProviderCredential(ctx, first); err != nil {
		t.Fatalf("replace first credential: %v", err)
	}
	second := memoryProviderCredential("credential-two", scope, "ciphertext-two", time.Now().UTC().Add(time.Minute))
	if err := store.ReplaceProviderCredential(ctx, second); err != nil {
		t.Fatalf("replace second credential: %v", err)
	}

	active, found, err := store.ActiveProviderCredential(ctx, scope)
	if err != nil {
		t.Fatalf("active provider credential: %v", err)
	}
	if !found || active.ID != "credential-two" || string(active.Sealed.Ciphertext) != "ciphertext-two" {
		t.Fatalf("unexpected active credential: found=%t credential=%+v", found, active)
	}
	if store.providerCreds["credential-one"].SupersededAt == nil {
		t.Fatalf("expected first credential to be superseded")
	}
}

func TestProviderCredentialRepositoryRejectsInvalidScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewStore()
	scope := memoryProviderCredentialScope()
	scope.ProviderProfileID = ""
	if err := store.ReplaceProviderCredential(ctx, memoryProviderCredential("credential-one", scope, "ciphertext-one", time.Now().UTC())); err != ports.ErrInvalidProviderCredential {
		t.Fatalf("expected invalid provider credential, got %v", err)
	}
}

func memoryProviderCredentialScope() ports.ProviderCredentialScope {
	return ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "profile-one",
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeAPIKey,
	}
}

func memoryProviderCredential(id string, scope ports.ProviderCredentialScope, ciphertext string, now time.Time) ports.ProviderCredentialRecord {
	return ports.ProviderCredentialRecord{
		ID:    id,
		Scope: scope,
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
