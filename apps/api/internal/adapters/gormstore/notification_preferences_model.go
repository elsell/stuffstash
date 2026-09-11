package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type notificationPreferencesModel struct {
	ID          string `gorm:"primaryKey;size:26"`
	TenantID    string `gorm:"not null;size:26;uniqueIndex:idx_notification_preferences_scope"`
	InventoryID string `gorm:"not null;size:26;uniqueIndex:idx_notification_preferences_scope"`
	PrincipalID string `gorm:"not null;size:128;uniqueIndex:idx_notification_preferences_scope"`
	Revision    int64  `gorm:"not null"`
	Settings    string `gorm:"type:jsonb;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (notificationPreferencesModel) TableName() string { return "notification_preferences" }
func (m notificationPreferencesModel) toRecord() (ports.NotificationPreferencesRecord, error) {
	var settings notification.Settings
	if err := json.Unmarshal([]byte(m.Settings), &settings); err != nil {
		return ports.NotificationPreferencesRecord{}, err
	}
	if err := settings.Validate(); err != nil {
		return ports.NotificationPreferencesRecord{}, err
	}
	return ports.NotificationPreferencesRecord{ID: m.ID, Scope: ports.NotificationScope{TenantID: tenant.ID(m.TenantID), InventoryID: inventory.InventoryID(m.InventoryID), PrincipalID: identity.PrincipalID(m.PrincipalID)}, Revision: m.Revision, Settings: settings.Clone(), CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}
