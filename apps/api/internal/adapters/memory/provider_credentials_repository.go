package memory

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) ReplaceProviderCredential(_ context.Context, credential ports.ProviderCredentialRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := credential.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	credential.CreatedAt = timeOrDefault(credential.CreatedAt, now)
	credential.UpdatedAt = now
	if err := validateProviderCredential(credential); err != nil {
		return err
	}
	for id, existingCredential := range s.providerCreds {
		if existingCredential.SupersededAt != nil || !sameProviderCredentialScope(existingCredential.Scope, credential.Scope) {
			continue
		}
		existingCredential.SupersededAt = &now
		existingCredential.UpdatedAt = now
		s.providerCreds[id] = existingCredential
	}
	s.providerCreds[credential.ID] = credential
	return nil
}

func (s *Store) ActiveProviderCredential(_ context.Context, scope ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := validateProviderCredentialScope(scope); err != nil {
		return ports.ProviderCredentialRecord{}, false, err
	}
	var latest ports.ProviderCredentialRecord
	found := false
	for _, credential := range s.providerCreds {
		if credential.SupersededAt != nil || !sameProviderCredentialScope(credential.Scope, scope) {
			continue
		}
		if !found || credential.CreatedAt.After(latest.CreatedAt) {
			latest = credential
			found = true
		}
	}
	return latest, found, nil
}

func (s *Store) ActiveProviderCredentialsExist(context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, credential := range s.providerCreds {
		if credential.SupersededAt == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) SupersedeActiveProviderCredential(_ context.Context, scope ports.ProviderCredentialScope, supersededAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateProviderCredentialScope(scope); err != nil {
		return err
	}
	if supersededAt.IsZero() {
		return ports.ErrInvalidProviderCredential
	}
	for id, credential := range s.providerCreds {
		if credential.SupersededAt != nil || !sameProviderCredentialScope(credential.Scope, scope) {
			continue
		}
		credential.SupersededAt = &supersededAt
		credential.UpdatedAt = supersededAt
		s.providerCreds[id] = credential
	}
	return nil
}

func validateProviderCredentialScope(scope ports.ProviderCredentialScope) error {
	if scope.TenantID.String() == "" || scope.ProviderProfileID == "" || scope.Capability == "" || scope.ProviderKind == "" || scope.Purpose == "" {
		return ports.ErrInvalidProviderCredential
	}
	switch scope.Capability {
	case ports.ProviderCapabilitySpeechToText, ports.ProviderCapabilityLanguageInference, ports.ProviderCapabilityTextToSpeech:
	default:
		return ports.ErrInvalidProviderCredential
	}
	switch scope.ProviderKind {
	case ports.ProviderKindGemini, ports.ProviderKindOpenAICompatible, ports.ProviderKindLocalHTTP:
	default:
		return ports.ErrInvalidProviderCredential
	}
	switch scope.Purpose {
	case ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer, ports.ProviderCredentialPurposeServerADC:
	default:
		return ports.ErrInvalidProviderCredential
	}
	return nil
}

func timeOrDefault(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}
