package notifications

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Service struct{ deps Dependencies }
type Dependencies struct {
	Assets      ports.AssetRepository
	Inbox       ports.NotificationInboxRepository
	Authorizer  ports.Authorizer
	Inventories ports.InventoryRepository
	Types       ports.CustomAssetTypeRepository
	Preferences ports.NotificationPreferencesRepository
	Audit       ports.AuditRepository
	IDs         ports.IDGenerator
	Clock       ports.Clock
	Observer    ports.Observer
}

func New(deps Dependencies) Service { return Service{deps: deps} }

type ScopeInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      audit.Source
	RequestID   string
}

func (i ScopeInput) Scope() ports.NotificationScope {
	return ports.NotificationScope{TenantID: i.TenantID, InventoryID: i.InventoryID, PrincipalID: i.Principal.ID}
}

func (s Service) access(ctx context.Context, input ScopeInput) error {
	if !input.Scope().Valid() || s.deps.Authorizer == nil || s.deps.Inventories == nil || s.deps.Preferences == nil {
		return apperrors.ErrInvalidInput
	}
	if err := s.deps.Authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionView, input.InventoryID); err != nil {
		return err
	}
	value, found, err := s.deps.Inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return err
	}
	if !found || !value.IsActive() {
		return apperrors.ErrNotFound
	}
	return nil
}
func (s Service) GetPreferences(ctx context.Context, input ScopeInput) (ports.NotificationPreferencesRecord, error) {
	if err := s.access(ctx, input); err != nil {
		return ports.NotificationPreferencesRecord{}, err
	}
	value, found, err := s.deps.Preferences.NotificationPreferences(ctx, input.Scope())
	if err != nil {
		return value, err
	}
	if !found {
		value = ports.NotificationPreferencesRecord{Scope: input.Scope(), Settings: notification.DefaultSettings("UTC")}
	}
	if err := appsupport.SaveReadAuditRecord(ctx, s.deps.Audit, s.deps.IDs, s.deps.Clock, s.auditInput(input, audit.ActionNotificationPreferencesViewed, input.InventoryID.String())); err != nil {
		return ports.NotificationPreferencesRecord{}, err
	}
	return value, nil
}
func (s Service) InitializePreferences(ctx context.Context, input ScopeInput, timezone string) (ports.NotificationPreferencesRecord, error) {
	if err := s.access(ctx, input); err != nil {
		return ports.NotificationPreferencesRecord{}, err
	}
	value, found, err := s.deps.Preferences.NotificationPreferences(ctx, input.Scope())
	if err != nil || found {
		return value, err
	}
	settings := notification.DefaultSettings(timezone)
	if settings.Validate() != nil {
		return value, apperrors.ErrInvalidInput
	}
	if s.deps.IDs == nil || s.deps.Clock == nil {
		return value, apperrors.ErrInvalidInput
	}
	now := s.deps.Clock.Now().UTC()
	value = ports.NotificationPreferencesRecord{ID: s.deps.IDs.NewID(), Scope: input.Scope(), Settings: settings, Revision: 1, CreatedAt: now, UpdatedAt: now}
	err = s.save(ctx, input, value, 0)
	if errors.Is(err, ports.ErrConflict) {
		existing, found, readErr := s.deps.Preferences.NotificationPreferences(ctx, input.Scope())
		if readErr != nil {
			return existing, readErr
		}
		if found {
			return existing, nil
		}
	}
	return value, err
}
func (s Service) UpdatePreferences(ctx context.Context, input ScopeInput, revision int64, defaults notification.ExpirationPreferences, timezone string, push bool) (ports.NotificationPreferencesRecord, error) {
	return s.update(ctx, input, revision, func(value *notification.Settings) error {
		value.Defaults = defaults
		value.Timezone = timezone
		value.PushEnabled = push
		return value.Validate()
	})
}
func (s Service) SetTypeOverride(ctx context.Context, input ScopeInput, revision int64, typeID string, override *notification.ExpirationPreferences) (ports.NotificationPreferencesRecord, error) {
	return s.update(ctx, input, revision, func(value *notification.Settings) error {
		if s.deps.Types == nil {
			return apperrors.ErrInvalidInput
		}
		kind, found, err := s.deps.Types.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, customfield.AssetTypeID(typeID))
		if err != nil {
			return err
		}
		if !found || !kind.IsActive() {
			return apperrors.ErrNotFound
		}
		id := notification.AssetTypeID(typeID)
		if override == nil {
			delete(value.Overrides, id)
		} else {
			value.Overrides[id] = *override
		}
		return value.Validate()
	})
}
func (s Service) update(ctx context.Context, input ScopeInput, revision int64, change func(*notification.Settings) error) (ports.NotificationPreferencesRecord, error) {
	if err := s.access(ctx, input); err != nil {
		return ports.NotificationPreferencesRecord{}, err
	}
	current, found, err := s.deps.Preferences.NotificationPreferences(ctx, input.Scope())
	if err != nil {
		return current, err
	}
	if !found || revision != current.Revision {
		return ports.NotificationPreferencesRecord{}, ports.ErrConflict
	}
	updated := current.Clone()
	if err := change(&updated.Settings); err != nil {
		if errors.Is(err, notification.ErrInvalidPreferences) {
			return updated, apperrors.ErrInvalidInput
		}
		return updated, err
	}
	if current.Settings.Equal(updated.Settings) {
		return current, nil
	}
	if s.deps.Clock == nil {
		return updated, apperrors.ErrInvalidInput
	}
	updated.Revision++
	updated.UpdatedAt = s.deps.Clock.Now().UTC()
	return updated, s.save(ctx, input, updated, current.Revision)
}
func (s Service) auditInput(input ScopeInput, action audit.Action, targetID string) appsupport.AuditRecordInput {
	return appsupport.AuditRecordInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: input.Source, RequestID: input.RequestID, Action: action, TargetType: audit.TargetNotificationPreferences, TargetID: targetID}
}
func (s Service) save(ctx context.Context, input ScopeInput, value ports.NotificationPreferencesRecord, expected int64) error {
	record, err := appsupport.NewAuditRecord(s.deps.IDs, s.deps.Clock, s.auditInput(input, audit.ActionNotificationPreferencesUpdated, value.ID))
	if err != nil {
		return err
	}
	if err := s.deps.Preferences.SaveNotificationPreferences(ctx, value, expected, record); err != nil {
		return err
	}
	if s.deps.Observer != nil {
		s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventNotificationPreferencesUpdated, Message: "notification preferences updated", Fields: map[string]string{"tenant_id": input.TenantID.String(), "inventory_id": input.InventoryID.String(), "principal_id": input.Principal.ID.String()}})
	}
	return nil
}
