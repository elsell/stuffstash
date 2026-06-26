package credentials

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestDatabaseProviderCredentialVaultPreparesSealedCredentialRecord(t *testing.T) {
	t.Parallel()

	sealer := &vaultSealer{}
	vault := NewDatabaseProviderCredentialVault(vaultRepository{}, sealer)
	now := time.Date(2026, 6, 26, 9, 0, 0, 0, time.UTC)

	record, err := vault.PrepareProviderCredential(context.Background(), ports.PrepareProviderCredentialInput{
		ID:        "credential-one",
		Scope:     vaultScope(),
		Raw:       []byte("raw-secret"),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("prepare credential: %v", err)
	}
	if record.ID != "credential-one" || record.Scope.ProviderProfileID != "profile-one" || !record.CreatedAt.Equal(now) || !record.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected prepared record: %+v", record)
	}
	if string(record.Sealed.Ciphertext) != "sealed:raw-secret" || len(sealer.sealedRaw) != 1 || string(sealer.sealedRaw[0]) != "raw-secret" {
		t.Fatalf("expected sealed raw material, got record %+v sealer %+v", record, sealer.sealedRaw)
	}
}

func TestDatabaseProviderCredentialVaultReadsActiveRawCredential(t *testing.T) {
	t.Parallel()

	scope := vaultScope()
	repository := vaultRepository{
		record: ports.ProviderCredentialRecord{
			ID:     "credential-one",
			Scope:  scope,
			Sealed: ports.SealedProviderCredential{KeyID: "key-one", Algorithm: ports.ProviderCredentialAlgorithmAES256GCM, Nonce: []byte("123456789012"), Ciphertext: []byte("sealed:raw-secret")},
		},
		found: true,
	}
	vault := NewDatabaseProviderCredentialVault(repository, &vaultSealer{})

	raw, found, err := vault.ActiveProviderCredentialMaterial(context.Background(), scope)
	if err != nil {
		t.Fatalf("read credential: %v", err)
	}
	if !found || string(raw) != "raw-secret" {
		t.Fatalf("expected active raw credential, found=%v raw=%q", found, string(raw))
	}
	raw[0] = 'X'
	if bytes.Equal(raw, repository.record.Sealed.Ciphertext) {
		t.Fatalf("expected raw credential copy independent of sealed storage")
	}
}

func TestDatabaseProviderCredentialVaultHandlesMissingCredential(t *testing.T) {
	t.Parallel()

	vault := NewDatabaseProviderCredentialVault(vaultRepository{}, &vaultSealer{})
	raw, found, err := vault.ActiveProviderCredentialMaterial(context.Background(), vaultScope())
	if err != nil {
		t.Fatalf("read missing credential: %v", err)
	}
	if found || raw != nil {
		t.Fatalf("expected missing credential, found=%v raw=%q", found, string(raw))
	}
}

func TestDatabaseProviderCredentialVaultRejectsInvalidCredentialBoundary(t *testing.T) {
	t.Parallel()

	scope := vaultScope()
	tests := map[string]struct {
		repository ports.ProviderCredentialRepository
		sealer     ports.ProviderCredentialSealer
	}{
		"nil repository": {
			sealer: &vaultSealer{},
		},
		"nil sealer": {
			repository: vaultRepository{},
		},
		"unseal failure": {
			repository: vaultRepository{record: ports.ProviderCredentialRecord{ID: "credential-one", Scope: scope}, found: true},
			sealer:     &vaultSealer{unsealErr: ports.ErrInvalidProviderCredential},
		},
		"empty raw material": {
			repository: vaultRepository{record: ports.ProviderCredentialRecord{ID: "credential-one", Scope: scope}, found: true},
			sealer:     &vaultSealer{raw: []byte{}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			vault := NewDatabaseProviderCredentialVault(tt.repository, tt.sealer)
			if _, _, err := vault.ActiveProviderCredentialMaterial(context.Background(), scope); err != ports.ErrInvalidProviderInput {
				t.Fatalf("expected invalid provider input, got %v", err)
			}
		})
	}
}

func vaultScope() ports.ProviderCredentialScope {
	return ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "profile-one",
		Capability:        ports.ProviderCapabilityLanguageInference,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeAPIKey,
	}
}

type vaultRepository struct {
	record ports.ProviderCredentialRecord
	found  bool
	err    error
}

func (r vaultRepository) ReplaceProviderCredential(context.Context, ports.ProviderCredentialRecord) error {
	return nil
}

func (r vaultRepository) ActiveProviderCredential(context.Context, ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	return r.record, r.found, r.err
}

func (r vaultRepository) ActiveProviderCredentialsExist(context.Context) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	return r.found, nil
}

func (r vaultRepository) SupersedeActiveProviderCredential(context.Context, ports.ProviderCredentialScope, time.Time) error {
	return nil
}

type vaultSealer struct {
	sealedRaw [][]byte
	raw       []byte
	unsealErr error
}

func (s *vaultSealer) SealProviderCredential(_ context.Context, _ ports.ProviderCredentialScope, raw []byte) (ports.SealedProviderCredential, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return ports.SealedProviderCredential{}, ports.ErrInvalidProviderCredential
	}
	s.sealedRaw = append(s.sealedRaw, append([]byte{}, raw...))
	return ports.SealedProviderCredential{
		KeyID:      "key-one",
		Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
		Nonce:      []byte("123456789012"),
		Ciphertext: []byte("sealed:" + string(raw)),
	}, nil
}

func (s *vaultSealer) UnsealProviderCredential(_ context.Context, _ ports.ProviderCredentialScope, sealed ports.SealedProviderCredential) ([]byte, error) {
	if s.unsealErr != nil {
		return nil, s.unsealErr
	}
	if s.raw != nil {
		return append([]byte{}, s.raw...), nil
	}
	raw, ok := bytes.CutPrefix(sealed.Ciphertext, []byte("sealed:"))
	if !ok {
		return nil, errors.New("invalid sealed material")
	}
	return append([]byte{}, raw...), nil
}
