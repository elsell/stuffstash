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
