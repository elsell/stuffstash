package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type notificationDeliveryModel struct {
	ID             string `gorm:"primaryKey;size:26"`
	TenantID       string `gorm:"not null;size:26"`
	InventoryID    string `gorm:"not null;size:26"`
	PrincipalID    string `gorm:"not null;size:128"`
	NotificationID string `gorm:"not null;size:26;uniqueIndex:idx_notification_delivery_device"`
	DeviceID       string `gorm:"not null;size:26;uniqueIndex:idx_notification_delivery_device"`
	DeviceRevision int64  `gorm:"not null"`
	CreatedAt      time.Time
	Status         string     `gorm:"not null;size:16;index:idx_notification_delivery_due"`
	Attempts       int        `gorm:"not null"`
	NextAttemptAt  *time.Time `gorm:"index:idx_notification_delivery_due"`
	LeaseUntil     *time.Time `gorm:"index:idx_notification_delivery_due"`
	Fence          string     `gorm:"not null;size:128"`
}

func (notificationDeliveryModel) TableName() string { return "notification_deliveries" }
func deliveryModel(value ports.NotificationDelivery) notificationDeliveryModel {
	m := notificationDeliveryModel{ID: value.ID, TenantID: value.Scope.TenantID.String(), InventoryID: value.Scope.InventoryID.String(), PrincipalID: value.Scope.PrincipalID.String(), NotificationID: value.NotificationID, DeviceID: value.DeviceID, DeviceRevision: value.DeviceRevision, CreatedAt: value.CreatedAt}
	m.setState(value.State)
	return m
}
func (m *notificationDeliveryModel) setState(state notification.DeliveryState) {
	m.Status = string(state.Status)
	m.Attempts = state.Attempts
	m.Fence = state.Fence
	m.NextAttemptAt = nil
	m.LeaseUntil = nil
	if !state.NextAttemptAt.IsZero() {
		value := state.NextAttemptAt
		m.NextAttemptAt = &value
	}
	if !state.LeaseUntil.IsZero() {
		value := state.LeaseUntil
		m.LeaseUntil = &value
	}
}
func (m notificationDeliveryModel) toDelivery() ports.NotificationDelivery {
	state := notification.DeliveryState{Status: notification.DeliveryStatus(m.Status), Attempts: m.Attempts, Fence: m.Fence}
	if m.NextAttemptAt != nil {
		state.NextAttemptAt = *m.NextAttemptAt
	}
	if m.LeaseUntil != nil {
		state.LeaseUntil = *m.LeaseUntil
	}
	return ports.NotificationDelivery{ID: m.ID, Scope: ports.NotificationScope{TenantID: tenant.ID(m.TenantID), InventoryID: inventory.InventoryID(m.InventoryID), PrincipalID: identity.PrincipalID(m.PrincipalID)}, NotificationID: m.NotificationID, DeviceID: m.DeviceID, DeviceRevision: m.DeviceRevision, CreatedAt: m.CreatedAt, State: state}
}
