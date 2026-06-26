package memory

import (
	"context"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) SaveProviderProfile(_ context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID(profile.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if _, exists := s.providerProfiles[profile.ID]; exists {
		return ports.ErrConflict
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.providerProfiles[profile.ID] = profile
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateProviderProfile(_ context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.providerProfiles[profile.ID]
	if !exists || existing.TenantID.String() != profile.TenantID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.providerProfiles[profile.ID] = profile
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) ReplaceProviderProfileCredential(_ context.Context, profile agentmodel.ProviderProfile, credential ports.ProviderCredentialRecord, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.providerProfiles[profile.ID]
	if !exists || existing.TenantID.String() != profile.TenantID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if err := validateProviderCredential(credential); err != nil {
		return err
	}
	if err := validateProviderCredentialMatchesProfile(profile, credential); err != nil {
		return err
	}
	for id, existingCredential := range s.providerCreds {
		if existingCredential.SupersededAt != nil || !sameProviderCredentialScope(existingCredential.Scope, credential.Scope) {
			continue
		}
		supersededAt := profile.UpdatedAt
		existingCredential.SupersededAt = &supersededAt
		existingCredential.UpdatedAt = supersededAt
		s.providerCreds[id] = existingCredential
	}
	s.providerProfiles[profile.ID] = profile
	s.providerCreds[credential.ID] = credential
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) ProviderProfileByID(_ context.Context, tenantID tenant.ID, profileID agentmodel.ProviderProfileID) (agentmodel.ProviderProfile, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, ok := s.providerProfiles[profileID]
	if !ok || profile.TenantID.String() != tenantID.String() {
		return agentmodel.ProviderProfile{}, false, nil
	}
	return profile, true, nil
}

func (s *Store) ListProviderProfiles(_ context.Context, tenantID tenant.ID) ([]agentmodel.ProviderProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profiles := []agentmodel.ProviderProfile{}
	for _, profile := range s.providerProfiles {
		if profile.TenantID.String() == tenantID.String() {
			profiles = append(profiles, profile)
		}
	}
	sort.Slice(profiles, func(left int, right int) bool {
		if profiles[left].CreatedAt.Equal(profiles[right].CreatedAt) {
			return profiles[left].ID.String() < profiles[right].ID.String()
		}
		return profiles[left].CreatedAt.Before(profiles[right].CreatedAt)
	})
	return profiles, nil
}

func validateProviderCredential(credential ports.ProviderCredentialRecord) error {
	if credential.ID == "" ||
		credential.Sealed.KeyID == "" ||
		credential.Sealed.Algorithm != ports.ProviderCredentialAlgorithmAES256GCM ||
		len(credential.Sealed.Nonce) != ports.ProviderCredentialAESGCMNonceBytes ||
		len(credential.Sealed.Ciphertext) == 0 ||
		credential.CreatedAt.IsZero() ||
		credential.UpdatedAt.IsZero() {
		return ports.ErrInvalidProviderCredential
	}
	if credential.Scope.TenantID.String() == "" || credential.Scope.ProviderProfileID == "" || credential.Scope.Capability == "" || credential.Scope.ProviderKind == "" || credential.Scope.Purpose == "" {
		return ports.ErrInvalidProviderCredential
	}
	switch credential.Scope.Capability {
	case ports.ProviderCapabilitySpeechToText, ports.ProviderCapabilityLanguageInference, ports.ProviderCapabilityTextToSpeech:
	default:
		return ports.ErrInvalidProviderCredential
	}
	switch credential.Scope.ProviderKind {
	case ports.ProviderKindGemini, ports.ProviderKindOpenAICompatible, ports.ProviderKindLocalHTTP:
	default:
		return ports.ErrInvalidProviderCredential
	}
	switch credential.Scope.Purpose {
	case ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer:
	default:
		return ports.ErrInvalidProviderCredential
	}
	return nil
}

func sameProviderCredentialScope(left ports.ProviderCredentialScope, right ports.ProviderCredentialScope) bool {
	return left.TenantID.String() == right.TenantID.String() &&
		left.ProviderProfileID == right.ProviderProfileID &&
		left.Capability == right.Capability &&
		left.ProviderKind == right.ProviderKind &&
		left.Purpose == right.Purpose
}

func validateProviderCredentialMatchesProfile(profile agentmodel.ProviderProfile, credential ports.ProviderCredentialRecord) error {
	if credential.Scope.TenantID.String() != profile.TenantID.String() ||
		credential.Scope.ProviderProfileID != profile.ID.String() ||
		string(credential.Scope.Capability) != profile.Capability.String() ||
		string(credential.Scope.ProviderKind) != profile.ProviderKind.String() {
		return ports.ErrInvalidProviderCredential
	}
	return nil
}
