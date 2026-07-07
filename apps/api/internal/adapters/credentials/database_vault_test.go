package credentials

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
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

func TestDatabaseImportJobSourceVaultTreatsExpiredSourceAsMissing(t *testing.T) {
	t.Parallel()

	scope := ports.ImportJobSourceScope{
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		JobID:       importjob.ID("job-one"),
	}
	repository := &vaultImportSourceRepository{
		record: ports.ImportJobSourceRecord{
			Scope:     scope,
			Sealed:    ports.SealedImportJobSource{KeyID: "key-one", Algorithm: ports.ProviderCredentialAlgorithmAES256GCM, Nonce: []byte("123456789012"), Ciphertext: []byte(`sealed:{"sourceType":"legacy_homebox","password":"secret"}`)},
			ExpiresAt: time.Now().Add(-time.Minute),
			CreatedAt: time.Now().Add(-2 * time.Minute),
			UpdatedAt: time.Now().Add(-2 * time.Minute),
		},
		found: true,
	}
	vault := NewDatabaseImportJobSourceVault(repository, &vaultSealer{})

	request, found, err := vault.ImportJobSourceRequest(context.Background(), scope)
	if err != nil {
		t.Fatalf("read expired import source: %v", err)
	}
	if found || request.Password != "" {
		t.Fatalf("expected expired import source to be missing, found=%v request=%+v", found, request)
	}
	if !repository.deleted {
		t.Fatalf("expected expired import source to be deleted")
	}
}

func TestDatabaseImportJobSourceVaultPreservesApplyAttachmentByteFlag(t *testing.T) {
	t.Parallel()

	scope := ports.ImportJobSourceScope{
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		JobID:       importjob.ID("job-one"),
	}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	repository := &vaultImportSourceRepository{}
	vault := NewDatabaseImportJobSourceVaultWithClock(repository, &vaultSealer{}, fixedClock{now: now})

	err := vault.StoreImportJobSource(context.Background(), scope, ports.ImportSourceRequest{
		SourceType:           importplan.SourceLegacyHomebox,
		BaseURL:              "https://homebox.example.test/api/v1",
		Username:             "owner@example.com",
		Password:             "secret",
		IncludeImages:        true,
		FetchAttachmentBytes: true,
		AllowPrivateNetwork:  true,
		MaxAttachmentBytes:   1234,
	}, now.Add(time.Minute), now)
	if err != nil {
		t.Fatalf("store import job source: %v", err)
	}

	request, found, err := vault.ImportJobSourceRequest(context.Background(), scope)
	if err != nil {
		t.Fatalf("read import job source: %v", err)
	}
	if !found {
		t.Fatalf("expected stored import source")
	}
	if !request.IncludeImages || !request.FetchAttachmentBytes || !request.AllowPrivateNetwork {
		t.Fatalf("expected import source flags to round trip, got %+v", request)
	}
	if request.SourceType != importplan.SourceLegacyHomebox || request.MaxAttachmentBytes != 1234 || request.Password != "secret" {
		t.Fatalf("unexpected import source request: %+v", request)
	}
}

func TestDatabaseImportJobSourceVaultReportsUnreadableStoredSourceMaterial(t *testing.T) {
	t.Parallel()

	scope := ports.ImportJobSourceScope{
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		JobID:       importjob.ID("job-one"),
	}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		sealer vaultSealer
		sealed ports.SealedImportJobSource
	}{
		"unseal failure": {
			sealer: vaultSealer{unsealErr: errors.New("ciphertext could not be opened")},
			sealed: ports.SealedImportJobSource{
				KeyID:      "key-one",
				Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
				Nonce:      []byte("123456789012"),
				Ciphertext: []byte(`sealed:{"sourceType":"legacy_homebox","password":"secret"}`),
			},
		},
		"invalid payload json": {
			sealed: ports.SealedImportJobSource{
				KeyID:      "key-one",
				Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
				Nonce:      []byte("123456789012"),
				Ciphertext: []byte(`sealed:not-json`),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repository := &vaultImportSourceRepository{
				record: ports.ImportJobSourceRecord{
					Scope:     scope,
					Sealed:    tt.sealed,
					ExpiresAt: now.Add(time.Minute),
					CreatedAt: now,
					UpdatedAt: now,
				},
				found: true,
			}
			vault := NewDatabaseImportJobSourceVaultWithClock(repository, &tt.sealer, fixedClock{now: now})

			request, found, err := vault.ImportJobSourceRequest(context.Background(), scope)
			if !errors.Is(err, ports.ErrImportJobSourceUnreadable) {
				t.Fatalf("expected unreadable import source error, got %v", err)
			}
			if found || request.Password != "" {
				t.Fatalf("expected unreadable import source to be unavailable, found=%v request=%+v", found, request)
			}
			if repository.deleted {
				t.Fatalf("did not expect unreadable source material to be deleted by read path")
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

func (s *vaultSealer) SealImportJobSource(_ context.Context, _ ports.ImportJobSourceScope, raw []byte) (ports.SealedImportJobSource, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return ports.SealedImportJobSource{}, ports.ErrInvalidProviderCredential
	}
	s.sealedRaw = append(s.sealedRaw, append([]byte{}, raw...))
	return ports.SealedImportJobSource{
		KeyID:      "key-one",
		Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
		Nonce:      []byte("123456789012"),
		Ciphertext: []byte("sealed:" + string(raw)),
	}, nil
}

func (s *vaultSealer) UnsealImportJobSource(_ context.Context, _ ports.ImportJobSourceScope, sealed ports.SealedImportJobSource) ([]byte, error) {
	if s.unsealErr != nil {
		return nil, s.unsealErr
	}
	raw, ok := bytes.CutPrefix(sealed.Ciphertext, []byte("sealed:"))
	if !ok {
		return nil, errors.New("invalid sealed material")
	}
	return append([]byte{}, raw...), nil
}

type vaultImportSourceRepository struct {
	record  ports.ImportJobSourceRecord
	found   bool
	deleted bool
}

func (r *vaultImportSourceRepository) ReplaceImportJobSource(_ context.Context, record ports.ImportJobSourceRecord) error {
	r.record = record
	r.found = true
	return nil
}

func (r *vaultImportSourceRepository) ImportJobSource(context.Context, ports.ImportJobSourceScope) (ports.ImportJobSourceRecord, bool, error) {
	return r.record, r.found, nil
}

func (r *vaultImportSourceRepository) DeleteImportJobSource(context.Context, ports.ImportJobSourceScope) (bool, error) {
	r.deleted = true
	return true, nil
}

func (r *vaultImportSourceRepository) DeleteExpiredImportJobSources(context.Context, time.Time) (int, error) {
	return 0, nil
}

func (r *vaultImportSourceRepository) DeleteVacuumableImportJobSources(context.Context, []importjob.Status, time.Time) ([]ports.ImportJobSourceScope, error) {
	return nil, nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}
