package bootstrap

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func validateProviderCredentialSealer(ctx context.Context, cfg config.Config, repository ports.ProviderCredentialRepository) error {
	configured := strings.TrimSpace(cfg.ProviderCredentialKeyID) != "" || strings.TrimSpace(cfg.ProviderCredentialKey) != ""
	activeCredentials := false
	if repository != nil {
		exists, err := repository.ActiveProviderCredentialsExist(ctx)
		if err != nil {
			return err
		}
		activeCredentials = exists
	}
	if !configured && !activeCredentials {
		return nil
	}
	if strings.TrimSpace(cfg.ProviderCredentialKeyID) == "" || strings.TrimSpace(cfg.ProviderCredentialKey) == "" {
		return errors.New("provider credential encryption key id and key are required")
	}
	if _, err := credentials.NewAESGCMSealerFromBase64(cfg.ProviderCredentialKeyID, cfg.ProviderCredentialKey); err != nil {
		return errors.New("provider credential encryption key is invalid")
	}
	return nil
}

func buildProviderCredentialSealer(cfg config.Config) (ports.ProviderCredentialSealer, error) {
	configured := strings.TrimSpace(cfg.ProviderCredentialKeyID) != "" || strings.TrimSpace(cfg.ProviderCredentialKey) != ""
	if !configured {
		return nil, nil
	}
	sealer, err := credentials.NewAESGCMSealerFromBase64(cfg.ProviderCredentialKeyID, cfg.ProviderCredentialKey)
	if err != nil {
		return nil, errors.New("provider credential encryption key is invalid")
	}
	return sealer, nil
}
