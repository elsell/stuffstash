package credentials

import (
	"bytes"
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type DatabaseProviderCredentialVault struct {
	repository ports.ProviderCredentialRepository
	sealer     ports.ProviderCredentialSealer
}

func NewDatabaseProviderCredentialVault(repository ports.ProviderCredentialRepository, sealer ports.ProviderCredentialSealer) DatabaseProviderCredentialVault {
	return DatabaseProviderCredentialVault{
		repository: repository,
		sealer:     sealer,
	}
}

func (v DatabaseProviderCredentialVault) PrepareProviderCredential(ctx context.Context, input ports.PrepareProviderCredentialInput) (ports.ProviderCredentialRecord, error) {
	if v.sealer == nil || input.ID == "" || input.CreatedAt.IsZero() || input.UpdatedAt.IsZero() || len(bytes.TrimSpace(input.Raw)) == 0 {
		return ports.ProviderCredentialRecord{}, ports.ErrInvalidProviderInput
	}
	sealed, err := v.sealer.SealProviderCredential(ctx, input.Scope, input.Raw)
	if err != nil {
		return ports.ProviderCredentialRecord{}, ports.ErrInvalidProviderInput
	}
	return ports.ProviderCredentialRecord{
		ID:        input.ID,
		Scope:     input.Scope,
		Sealed:    sealed,
		CreatedAt: input.CreatedAt,
		UpdatedAt: input.UpdatedAt,
	}, nil
}

func (v DatabaseProviderCredentialVault) ActiveProviderCredentialMaterial(ctx context.Context, scope ports.ProviderCredentialScope) ([]byte, bool, error) {
	material, found, err := v.ActiveVersionedProviderCredential(ctx, scope)
	return material.Raw, found, err
}

func (v DatabaseProviderCredentialVault) ActiveVersionedProviderCredential(ctx context.Context, scope ports.ProviderCredentialScope) (ports.VersionedProviderCredentialMaterial, bool, error) {
	if v.repository == nil || v.sealer == nil {
		return ports.VersionedProviderCredentialMaterial{}, false, ports.ErrInvalidProviderInput
	}
	record, found, err := v.repository.ActiveProviderCredential(ctx, scope)
	if err != nil {
		return ports.VersionedProviderCredentialMaterial{}, false, err
	}
	if !found {
		return ports.VersionedProviderCredentialMaterial{}, false, nil
	}
	if record.Scope != scope || record.SupersededAt != nil || strings.TrimSpace(record.ID) == "" {
		return ports.VersionedProviderCredentialMaterial{}, false, ports.ErrInvalidProviderInput
	}
	raw, err := v.sealer.UnsealProviderCredential(ctx, scope, record.Sealed)
	if err != nil || len(bytes.TrimSpace(raw)) == 0 {
		return ports.VersionedProviderCredentialMaterial{}, false, ports.ErrInvalidProviderInput
	}
	return ports.VersionedProviderCredentialMaterial{VersionID: record.ID, Raw: append([]byte{}, raw...)}, true, nil
}
