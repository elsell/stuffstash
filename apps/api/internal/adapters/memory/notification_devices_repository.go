package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) NotificationDeviceByID(ctx context.Context, scope ports.NotificationScope, id string) (ports.NotificationDevice, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.NotificationDevice{}, false, err
	}
	if !scope.Valid() || id == "" {
		return ports.NotificationDevice{}, false, ports.ErrForbidden
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.notificationDevices[id]
	if !found || value.Scope != scope {
		return ports.NotificationDevice{}, false, nil
	}
	return value, true, nil
}
func (s *Store) NotificationDeviceByInstallation(ctx context.Context, scope ports.NotificationScope, installation string) (ports.NotificationDevice, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.NotificationDevice{}, false, err
	}
	if !scope.Valid() || installation == "" {
		return ports.NotificationDevice{}, false, ports.ErrForbidden
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, value := range s.notificationDevices {
		if value.Scope == scope && value.InstallationID == installation {
			return value, true, nil
		}
	}
	return ports.NotificationDevice{}, false, nil
}
func (s *Store) ListNotificationDevices(ctx context.Context, scope ports.NotificationScope, after string, limit int) ([]ports.NotificationDevice, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !scope.Valid() {
		return nil, ports.ErrForbidden
	}
	if limit < 1 || limit > 1000 {
		return nil, ports.ErrInvalidProviderInput
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := []ports.NotificationDevice{}
	for _, value := range s.notificationDevices {
		if value.Scope == scope && value.Active && value.ID > after {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	if len(values) > limit {
		values = values[:limit]
	}
	return values, nil
}
func (s *Store) SaveNotificationDevice(ctx context.Context, value ports.NotificationDevice, expected int64, record audit.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !value.Valid() || expected < 0 || value.Revision != expected+1 {
		return ports.ErrInvalidProviderInput
	}
	if record.TenantID.String() != value.Scope.TenantID.String() || record.InventoryID.String() != value.Scope.InventoryID.String() || record.PrincipalID.String() != value.Scope.PrincipalID.String() {
		return ports.ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inventory, found := s.inventories[value.Scope.InventoryID]
	if !found || inventory.TenantID.String() != value.Scope.TenantID.String() {
		return ports.ErrForbidden
	}
	current, found := s.notificationDevices[value.ID]
	if expected == 0 && found || expected > 0 && (!found || current.Scope != value.Scope || current.Revision != expected || current.InstallationID != value.InstallationID || !current.CreatedAt.Equal(value.CreatedAt)) {
		return ports.ErrConflict
	}
	for _, existing := range s.notificationDevices {
		if existing.ID == value.ID {
			continue
		}
		if existing.Scope == value.Scope && existing.InstallationID == value.InstallationID {
			return ports.ErrConflict
		}
		if value.Active && existing.Active && existing.Transport == value.Transport && existing.Token == value.Token && existing.Scope.PrincipalID != value.Scope.PrincipalID {
			return ports.ErrConflict
		}
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrConflict
	}
	if s.notificationDevices == nil {
		s.notificationDevices = map[string]ports.NotificationDevice{}
	}
	s.notificationDevices[value.ID] = value
	s.auditRecords[record.ID] = record
	return nil
}
