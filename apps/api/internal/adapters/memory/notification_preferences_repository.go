package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) NotificationPreferences(_ context.Context, scope ports.NotificationScope) (ports.NotificationPreferencesRecord, bool, error) {
	if !scope.Valid() {
		return ports.NotificationPreferencesRecord{}, false, ports.ErrForbidden
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.notificationPreferences[scope]
	return value.Clone(), found, nil
}
func (s *Store) SaveNotificationPreferences(_ context.Context, value ports.NotificationPreferencesRecord, expected int64, record audit.Record) error {
	if !value.Scope.Valid() || value.ID == "" || value.Settings.Validate() != nil || value.Revision != expected+1 {
		return ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inventory, found := s.inventories[value.Scope.InventoryID]
	if !found || inventory.TenantID.String() != value.Scope.TenantID.String() {
		return ports.ErrForbidden
	}
	current, found := s.notificationPreferences[value.Scope]
	if (expected == 0 && found) || (expected != 0 && (!found || current.Revision != expected || current.ID != value.ID)) {
		return ports.ErrConflict
	}
	if record.TenantID.String() != value.Scope.TenantID.String() || record.InventoryID.String() != value.Scope.InventoryID.String() || record.PrincipalID.String() != value.Scope.PrincipalID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrConflict
	}
	if s.notificationPreferences == nil {
		s.notificationPreferences = map[ports.NotificationScope]ports.NotificationPreferencesRecord{}
	}
	s.notificationPreferences[value.Scope] = value.Clone()
	s.auditRecords[record.ID] = record
	return nil
}
func (s *Store) ListNotificationRecipients(_ context.Context, afterID string, limit int) ([]ports.NotificationPreferencesRecord, error) {
	if limit < 1 || limit > 1000 {
		return nil, ports.ErrInvalidProviderInput
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []ports.NotificationPreferencesRecord{}
	for _, value := range s.notificationPreferences {
		if value.ID > afterID {
			result = append(result, value.Clone())
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
