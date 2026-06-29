package memory

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) VoiceProviderConfiguration(_ context.Context, tenantID tenant.ID) (ports.VoiceProviderConfigurationRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, found := s.voiceConfigs[tenantID]
	return record, found, nil
}

func (s *Store) SaveVoiceProviderConfiguration(_ context.Context, record ports.VoiceProviderConfigurationRecord, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[record.TenantID]; !exists {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.voiceConfigs[record.TenantID] = record
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}
