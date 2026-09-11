package gormstore

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type notificationDeviceModel struct {
	ID             string `gorm:"primaryKey;size:26"`
	TenantID       string `gorm:"not null;size:26;uniqueIndex:idx_notification_device_installation;index:idx_notification_device_scope"`
	InventoryID    string `gorm:"not null;size:26;uniqueIndex:idx_notification_device_installation;index:idx_notification_device_scope"`
	PrincipalID    string `gorm:"not null;size:128;uniqueIndex:idx_notification_device_installation;index:idx_notification_device_scope"`
	InstallationID string `gorm:"not null;size:128;uniqueIndex:idx_notification_device_installation"`
	Transport      string `gorm:"not null;size:8"`
	Token          string `gorm:"not null;size:4096"`
	TokenKey       string `gorm:"not null;size:64;index:idx_notification_device_token"`
	Active         bool   `gorm:"not null"`
	Revision       int64  `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (notificationDeviceModel) TableName() string { return "notification_devices" }

type notificationTokenLockModel struct {
	Key string `gorm:"primaryKey;size:64"`
}

func (notificationTokenLockModel) TableName() string { return "notification_device_token_locks" }
func notificationDeviceScope(scope ports.NotificationScope) notificationDeviceModel {
	return notificationDeviceModel{TenantID: scope.TenantID.String(), InventoryID: scope.InventoryID.String(), PrincipalID: scope.PrincipalID.String()}
}
func deviceModel(value ports.NotificationDevice) notificationDeviceModel {
	model := notificationDeviceScope(value.Scope)
	model.ID = value.ID
	model.InstallationID = value.InstallationID
	model.Transport = string(value.Transport)
	model.Token = value.Token.Secret()
	sum := sha256.Sum256([]byte(model.Transport + "\x00" + model.Token))
	model.TokenKey = hex.EncodeToString(sum[:])
	model.Active = value.Active
	model.Revision = value.Revision
	model.CreatedAt = value.CreatedAt
	model.UpdatedAt = value.UpdatedAt
	return model
}
func (m notificationDeviceModel) toDevice() (ports.NotificationDevice, error) {
	token, err := notification.ParseDeviceToken(m.Token)
	if err != nil {
		return ports.NotificationDevice{}, ports.ErrInvalidProviderInput
	}
	value := ports.NotificationDevice{ID: m.ID, Scope: ports.NotificationScope{TenantID: tenant.ID(m.TenantID), InventoryID: inventory.InventoryID(m.InventoryID), PrincipalID: identity.PrincipalID(m.PrincipalID)}, InstallationID: m.InstallationID, Transport: notification.PushTransport(m.Transport), Token: token, Active: m.Active, Revision: m.Revision, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
	if !value.Valid() {
		return ports.NotificationDevice{}, ports.ErrInvalidProviderInput
	}
	return value, nil
}
