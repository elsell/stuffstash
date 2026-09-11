package notifications

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type RegisterDeviceInput struct {
	InstallationID string
	Transport      notification.PushTransport
	Token          notification.DeviceToken
	Revision       int64
}

func (s Service) RegisterDevice(ctx context.Context, scope ScopeInput, input RegisterDeviceInput) (ports.NotificationDevice, error) {
	if err := s.access(ctx, scope); err != nil {
		return ports.NotificationDevice{}, err
	}
	if s.deps.Devices == nil || s.deps.PushTokens == nil || s.deps.IDs == nil || s.deps.Clock == nil || input.Revision < 0 || input.InstallationID == "" || len(input.InstallationID) > 128 || !input.Transport.Valid() || input.Token.Empty() {
		return ports.NotificationDevice{}, apperrors.ErrInvalidInput
	}
	normalized, err := s.deps.PushTokens.NormalizeDeviceToken(ctx, input.Transport, input.Token)
	if err != nil {
		return ports.NotificationDevice{}, apperrors.ErrInvalidInput
	}
	input.Token = normalized
	current, found, err := s.deps.Devices.NotificationDeviceByInstallation(ctx, scope.Scope(), input.InstallationID)
	if err != nil {
		return ports.NotificationDevice{}, err
	}
	same := found && current.Active && current.Transport == input.Transport && current.Token == input.Token
	if same && (input.Revision == 0 || input.Revision == current.Revision) {
		return current, nil
	}
	if found && input.Revision != current.Revision || !found && input.Revision != 0 {
		return ports.NotificationDevice{}, ports.ErrConflict
	}
	now := s.deps.Clock.Now().UTC()
	if !found {
		current = ports.NotificationDevice{ID: s.deps.IDs.NewID(), Scope: scope.Scope(), InstallationID: input.InstallationID, CreatedAt: now}
	}
	expected := current.Revision
	current.Revision++
	current.UpdatedAt = now
	current.Active = true
	current.Transport = input.Transport
	current.Token = input.Token
	if err := s.saveDevice(ctx, scope, current, expected, audit.ActionNotificationDeviceUpdated, ports.EventNotificationDeviceUpdated); err != nil {
		return ports.NotificationDevice{}, err
	}
	return current, nil
}
func (s Service) GetDeviceByInstallation(ctx context.Context, scope ScopeInput, installation string) (ports.NotificationDevice, error) {
	if !scope.Scope().Valid() {
		return ports.NotificationDevice{}, apperrors.ErrInvalidInput
	}
	if s.deps.Devices == nil || installation == "" {
		return ports.NotificationDevice{}, apperrors.ErrInvalidInput
	}
	value, found, err := s.deps.Devices.NotificationDeviceByInstallation(ctx, scope.Scope(), installation)
	if err != nil {
		return ports.NotificationDevice{}, err
	}
	if !found {
		return ports.NotificationDevice{}, apperrors.ErrNotFound
	}
	input := s.auditInput(scope, audit.ActionNotificationDeviceViewed, value.ID)
	input.TargetType = audit.TargetNotificationDevice
	if err := appsupport.SaveReadAuditRecord(ctx, s.deps.Audit, s.deps.IDs, s.deps.Clock, input); err != nil {
		return ports.NotificationDevice{}, err
	}
	return value, nil
}
func (s Service) RevokeDevice(ctx context.Context, scope ScopeInput, id string, revision int64) (ports.NotificationDevice, error) {
	if !scope.Scope().Valid() {
		return ports.NotificationDevice{}, apperrors.ErrInvalidInput
	}
	if s.deps.Devices == nil || s.deps.Clock == nil || id == "" || revision < 1 {
		return ports.NotificationDevice{}, apperrors.ErrInvalidInput
	}
	value, found, err := s.deps.Devices.NotificationDeviceByID(ctx, scope.Scope(), id)
	if err != nil {
		return ports.NotificationDevice{}, err
	}
	if !found {
		return ports.NotificationDevice{}, apperrors.ErrNotFound
	}
	if value.Revision != revision {
		return ports.NotificationDevice{}, ports.ErrConflict
	}
	if !value.Active {
		return value, nil
	}
	value.Revision++
	value.Active = false
	value.UpdatedAt = s.deps.Clock.Now().UTC()
	if err := s.saveDevice(ctx, scope, value, revision, audit.ActionNotificationDeviceRevoked, ports.EventNotificationDeviceRevoked); err != nil {
		return ports.NotificationDevice{}, err
	}
	return value, nil
}
func (s Service) saveDevice(ctx context.Context, scope ScopeInput, value ports.NotificationDevice, expected int64, action audit.Action, event ports.EventName) error {
	input := s.auditInput(scope, action, value.ID)
	input.TargetType = audit.TargetNotificationDevice
	record, err := appsupport.NewAuditRecord(s.deps.IDs, s.deps.Clock, input)
	if err != nil {
		return err
	}
	if err := s.deps.Devices.SaveNotificationDevice(ctx, value, expected, record); err != nil {
		return err
	}
	if s.deps.Observer != nil {
		s.deps.Observer.Record(ctx, ports.Event{Name: event, Message: "notification device changed", Fields: map[string]string{"tenant_id": scope.TenantID.String(), "inventory_id": scope.InventoryID.String(), "principal_id": scope.Principal.ID.String(), "device_id": value.ID}})
	}
	return nil
}
