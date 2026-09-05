package credentials

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type rotatingVaultRepository struct {
	ports.ProviderCredentialRepository
	records []ports.ProviderCredentialRecord
	reads   int
}

func (r *rotatingVaultRepository) ActiveProviderCredential(_ context.Context, scope ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	if r.reads >= len(r.records) {
		return ports.ProviderCredentialRecord{}, false, nil
	}
	value := r.records[r.reads]
	r.reads++
	return value, true, nil
}
func versionedVaultRecord(id, raw string) ports.ProviderCredentialRecord {
	return ports.ProviderCredentialRecord{ID: id, Scope: vaultScope(), Sealed: ports.SealedProviderCredential{Ciphertext: []byte("sealed:" + raw)}}
}
func TestVersionedVaultPairsMaterialWithSingleReadVersion(t *testing.T) {
	repository := &rotatingVaultRepository{records: []ports.ProviderCredentialRecord{versionedVaultRecord("one", "first"), versionedVaultRecord("two", "second")}}
	vault := NewDatabaseProviderCredentialVault(repository, &vaultSealer{})
	first, found, err := vault.ActiveVersionedProviderCredential(context.Background(), vaultScope())
	if err != nil || !found || first.VersionID != "one" || string(first.Raw) != "first" || repository.reads != 1 {
		t.Fatal("credential version and material not paired")
	}
	first.Raw[0] = 'X'
	second, found, err := vault.ActiveVersionedProviderCredential(context.Background(), vaultScope())
	if err != nil || !found || second.VersionID != "two" || string(second.Raw) != "second" || repository.reads != 2 {
		t.Fatal("credential replacement not reflected")
	}
	if string(repository.records[0].Sealed.Ciphertext) != "sealed:first" {
		t.Fatal("caller mutated stored credential")
	}
}
func TestVersionedVaultRejectsMismatchedOrRetiredRecords(t *testing.T) {
	for _, scenario := range []string{"tenant", "profile", "capability", "kind", "purpose", "superseded", "blank version"} {
		t.Run(scenario, func(t *testing.T) {
			record := versionedVaultRecord("one", "secret")
			switch scenario {
			case "tenant":
				record.Scope.TenantID = "outside"
			case "profile":
				record.Scope.ProviderProfileID = "outside"
			case "capability":
				record.Scope.Capability = ports.ProviderCapabilityTextToSpeech
			case "kind":
				record.Scope.ProviderKind = ports.ProviderKindLocalHTTP
			case "purpose":
				record.Scope.Purpose = ports.ProviderCredentialPurposeServerADC
			case "superseded":
				now := time.Now()
				record.SupersededAt = &now
			case "blank version":
				record.ID = " "
			}
			vault := NewDatabaseProviderCredentialVault(vaultRepository{record: record, found: true}, &vaultSealer{})
			material, found, err := vault.ActiveVersionedProviderCredential(context.Background(), vaultScope())
			if err == nil || found || material.VersionID != "" || len(material.Raw) != 0 {
				t.Fatal("invalid credential exposed")
			}
			raw, found, err := vault.ActiveProviderCredentialMaterial(context.Background(), vaultScope())
			if err == nil || found || len(raw) != 0 {
				t.Fatal("legacy read bypassed credential validation")
			}
		})
	}
}
func TestVersionedVaultMissingCredentialHasNoIdentity(t *testing.T) {
	vault := NewDatabaseProviderCredentialVault(vaultRepository{}, &vaultSealer{})
	material, found, err := vault.ActiveVersionedProviderCredential(context.Background(), vaultScope())
	if err != nil || found || material.VersionID != "" || len(material.Raw) != 0 {
		t.Fatal("missing credential acquired identity")
	}
}
