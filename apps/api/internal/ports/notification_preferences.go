package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"time"
)

type NotificationScope struct {
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	PrincipalID identity.PrincipalID
}

func (s NotificationScope) Valid() bool {
	return s.TenantID != "" && s.InventoryID != "" && s.PrincipalID != ""
}

type NotificationPreferencesRecord struct {
	ID        string
	Scope     NotificationScope
	Revision  int64
	Settings  notification.Settings
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r NotificationPreferencesRecord) Clone() NotificationPreferencesRecord {
	r.Settings = r.Settings.Clone()
	return r
}

type NotificationPreferencesRepository interface {
	NotificationPreferences(ctx context.Context, scope NotificationScope) (NotificationPreferencesRecord, bool, error)
	SaveNotificationPreferences(ctx context.Context, record NotificationPreferencesRecord, expectedRevision int64, auditRecord audit.Record) error
	// Operational discovery is permitted to enumerate registered recipient scopes.
	ListNotificationRecipients(ctx context.Context, afterID string, limit int) ([]NotificationPreferencesRecord, error)
}
