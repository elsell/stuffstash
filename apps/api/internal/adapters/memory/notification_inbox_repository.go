package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) InsertNotification(_ context.Context, value ports.NotificationRecord, record audit.Record) (ports.NotificationRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.insertNotificationLocked(value, record)
}
func (s *Store) insertNotificationLocked(value ports.NotificationRecord, record audit.Record) (ports.NotificationRecord, bool, error) {
	if !value.Valid() || value.ReadAt != nil {
		return ports.NotificationRecord{}, false, ports.ErrInvalidProviderInput
	}
	if !notificationAuditScopeMatches(value.Scope, record) {
		return ports.NotificationRecord{}, false, ports.ErrForbidden
	}
	inv, found := s.inventories[value.Scope.InventoryID]
	if !found || inv.TenantID.String() != value.Scope.TenantID.String() {
		return ports.NotificationRecord{}, false, ports.ErrForbidden
	}
	for _, existing := range s.notificationInbox {
		if existing.Scope == value.Scope && existing.Milestone == value.Milestone {
			return existing.Clone(), false, nil
		}
	}
	if _, exists := s.notificationInbox[value.ID]; exists {
		return ports.NotificationRecord{}, false, ports.ErrConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.NotificationRecord{}, false, ports.ErrConflict
	}
	if s.notificationInbox == nil {
		s.notificationInbox = map[string]ports.NotificationRecord{}
	}
	s.notificationInbox[value.ID] = value.Clone()
	s.auditRecords[record.ID] = record
	return value.Clone(), true, nil
}
func notificationAuditScopeMatches(scope ports.NotificationScope, record audit.Record) bool {
	return scope.Valid() && record.TenantID.String() == scope.TenantID.String() && record.InventoryID.String() == scope.InventoryID.String() && record.PrincipalID.String() == scope.PrincipalID.String()
}
func (s *Store) NotificationByID(_ context.Context, scope ports.NotificationScope, id string) (ports.NotificationRecord, bool, error) {
	if id == "" {
		return ports.NotificationRecord{}, false, ports.ErrInvalidProviderInput
	}
	if !scope.Valid() {
		return ports.NotificationRecord{}, false, ports.ErrForbidden
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.notificationInbox[id]
	if !found || value.Scope != scope {
		return ports.NotificationRecord{}, false, nil
	}
	return value.Clone(), true, nil
}
func (s *Store) ListNotifications(_ context.Context, scope ports.NotificationScope, beforeID string, limit int) ([]ports.NotificationRecord, error) {
	if !scope.Valid() {
		return nil, ports.ErrForbidden
	}
	if limit < 1 || limit > 1000 {
		return nil, ports.ErrInvalidProviderInput
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []ports.NotificationRecord{}
	for _, value := range s.notificationInbox {
		if value.Scope == scope && (beforeID == "" || value.ID < beforeID) {
			result = append(result, value.Clone())
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID > result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
func (s *Store) MarkNotificationRead(ctx context.Context, scope ports.NotificationScope, id string, at time.Time, record audit.Record) (bool, error) {
	if at.IsZero() {
		return false, ports.ErrInvalidProviderInput
	}
	at = at.UTC()
	return s.setNotificationRead(ctx, scope, id, &at, record)
}
func (s *Store) MarkNotificationUnread(ctx context.Context, scope ports.NotificationScope, id string, record audit.Record) (bool, error) {
	return s.setNotificationRead(ctx, scope, id, nil, record)
}
func (s *Store) setNotificationRead(ctx context.Context, scope ports.NotificationScope, id string, at *time.Time, record audit.Record) (bool, error) {
	if id == "" {
		return false, ports.ErrInvalidProviderInput
	}
	if !notificationAuditScopeMatches(scope, record) {
		return false, ports.ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, found := s.notificationInbox[id]
	if !found || value.Scope != scope {
		return false, ports.ErrForbidden
	}
	if (value.ReadAt != nil) == (at != nil) {
		return false, nil
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return false, ports.ErrConflict
	}
	value.ReadAt = at
	s.notificationInbox[id] = value
	s.auditRecords[record.ID] = record
	return true, nil
}
