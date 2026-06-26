package credentials

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAESGCMSealerRoundTripsCredentialWithScopeBinding(t *testing.T) {
	t.Parallel()

	sealer := newTestSealer(t)
	scope := testCredentialScope()

	sealed, err := sealer.SealProviderCredential(context.Background(), scope, []byte("secret-token"))
	if err != nil {
		t.Fatalf("seal credential: %v", err)
	}
	if sealed.KeyID != "local-key" || sealed.Algorithm != AES256GCMAlgorithm || len(sealed.Nonce) != 12 || bytes.Contains(sealed.Ciphertext, []byte("secret-token")) {
		t.Fatalf("unexpected sealed credential metadata: %+v", sealed)
	}

	raw, err := sealer.UnsealProviderCredential(context.Background(), scope, sealed)
	if err != nil {
		t.Fatalf("unseal credential: %v", err)
	}
	if string(raw) != "secret-token" {
		t.Fatalf("unexpected raw credential %q", string(raw))
	}
}

func TestAESGCMSealerRejectsWrongScope(t *testing.T) {
	t.Parallel()

	sealer := newTestSealer(t)
	scope := testCredentialScope()
	sealed, err := sealer.SealProviderCredential(context.Background(), scope, []byte("secret-token"))
	if err != nil {
		t.Fatalf("seal credential: %v", err)
	}

	cases := map[string]ports.ProviderCredentialScope{
		"tenant": {
			TenantID:          tenant.ID("tenant-other"),
			ProviderProfileID: scope.ProviderProfileID,
			Capability:        scope.Capability,
			ProviderKind:      scope.ProviderKind,
			Purpose:           scope.Purpose,
		},
		"profile": {
			TenantID:          scope.TenantID,
			ProviderProfileID: "profile-other",
			Capability:        scope.Capability,
			ProviderKind:      scope.ProviderKind,
			Purpose:           scope.Purpose,
		},
		"capability": {
			TenantID:          scope.TenantID,
			ProviderProfileID: scope.ProviderProfileID,
			Capability:        ports.ProviderCapabilityTextToSpeech,
			ProviderKind:      scope.ProviderKind,
			Purpose:           scope.Purpose,
		},
		"provider kind": {
			TenantID:          scope.TenantID,
			ProviderProfileID: scope.ProviderProfileID,
			Capability:        scope.Capability,
			ProviderKind:      ports.ProviderKindOpenAICompatible,
			Purpose:           scope.Purpose,
		},
		"purpose": {
			TenantID:          scope.TenantID,
			ProviderProfileID: scope.ProviderProfileID,
			Capability:        scope.Capability,
			ProviderKind:      scope.ProviderKind,
			Purpose:           ports.ProviderCredentialPurposeOAuthBearer,
		},
	}
	for name, wrongScope := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := sealer.UnsealProviderCredential(context.Background(), wrongScope, sealed); !errors.Is(err, ports.ErrInvalidProviderCredential) {
				t.Fatalf("expected invalid credential error, got %v", err)
			}
		})
	}
}

func TestAESGCMSealerRejectsMisconfigurationAndTampering(t *testing.T) {
	t.Parallel()

	if _, err := NewAESGCMSealer(AESGCMSealerConfig{KeyID: "local-key", Key: []byte("short")}); !errors.Is(err, ports.ErrInvalidProviderCredential) {
		t.Fatalf("expected short key rejection, got %v", err)
	}
	if _, err := NewAESGCMSealer(AESGCMSealerConfig{Key: bytes.Repeat([]byte{1}, 32)}); !errors.Is(err, ports.ErrInvalidProviderCredential) {
		t.Fatalf("expected missing key id rejection, got %v", err)
	}
	encoded := base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32))
	if _, err := NewAESGCMSealerFromBase64("local-key", encoded); err != nil {
		t.Fatalf("expected raw base64 key to be accepted: %v", err)
	}

	sealer := newTestSealer(t)
	scope := testCredentialScope()
	sealed, err := sealer.SealProviderCredential(context.Background(), scope, []byte("secret-token"))
	if err != nil {
		t.Fatalf("seal credential: %v", err)
	}
	sealed.Ciphertext[0] ^= 0xff
	if _, err := sealer.UnsealProviderCredential(context.Background(), scope, sealed); !errors.Is(err, ports.ErrInvalidProviderCredential) {
		t.Fatalf("expected tampered ciphertext rejection, got %v", err)
	}
}

func TestAESGCMSealerRejectsControlCharactersInScope(t *testing.T) {
	t.Parallel()

	sealer := newTestSealer(t)
	scope := testCredentialScope()
	scope.ProviderProfileID = "profile\x00google"
	if _, err := sealer.SealProviderCredential(context.Background(), scope, []byte("secret-token")); !errors.Is(err, ports.ErrInvalidProviderCredential) {
		t.Fatalf("expected control character scope rejection, got %v", err)
	}
}

func newTestSealer(t *testing.T) AESGCMSealer {
	t.Helper()

	sealer, err := NewAESGCMSealer(AESGCMSealerConfig{
		KeyID:  "local-key",
		Key:    bytes.Repeat([]byte{1}, 32),
		Random: bytes.NewReader(bytes.Repeat([]byte{7}, 1024)),
	})
	if err != nil {
		t.Fatalf("new sealer: %v", err)
	}
	return sealer
}

func testCredentialScope() ports.ProviderCredentialScope {
	return ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "profile-google",
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeAPIKey,
	}
}
