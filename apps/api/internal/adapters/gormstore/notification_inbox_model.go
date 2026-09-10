package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type notificationInboxModel struct {
	ID                  string `gorm:"primaryKey;size:26"`
	TenantID            string `gorm:"not null;size:26;uniqueIndex:idx_notification_milestone;index:idx_notification_inbox_scope"`
	InventoryID         string `gorm:"not null;size:26;uniqueIndex:idx_notification_milestone;index:idx_notification_inbox_scope"`
	PrincipalID         string `gorm:"not null;size:128;uniqueIndex:idx_notification_milestone;index:idx_notification_inbox_scope"`
	AssetID             string `gorm:"not null;size:26;uniqueIndex:idx_notification_milestone"`
	ExpirationDate      string `gorm:"not null;size:10;uniqueIndex:idx_notification_milestone"`
	ExpirationPrecision string `gorm:"not null;size:8;uniqueIndex:idx_notification_milestone"`
	Kind                string `gorm:"not null;size:16;uniqueIndex:idx_notification_milestone"`
	CreatedAt           time.Time
	ReadAt              *time.Time
}

func (notificationInboxModel) TableName() string { return "notification_inbox" }
func notificationInboxScope(scope ports.NotificationScope) notificationInboxModel {
	return notificationInboxModel{TenantID: scope.TenantID.String(), InventoryID: scope.InventoryID.String(), PrincipalID: scope.PrincipalID.String()}
}
func (m notificationInboxModel) toRecord() (ports.NotificationRecord, error) {
	date, err := expirationdate.ParseDate(m.ExpirationDate, expirationdate.Precision(m.ExpirationPrecision))
	if err != nil {
		return ports.NotificationRecord{}, err
	}
	value := ports.NotificationRecord{ID: m.ID, Scope: ports.NotificationScope{TenantID: tenant.ID(m.TenantID), InventoryID: inventory.InventoryID(m.InventoryID), PrincipalID: identity.PrincipalID(m.PrincipalID)}, Milestone: notification.Milestone{AssetID: m.AssetID, Date: date, Kind: notification.MilestoneKind(m.Kind)}, CreatedAt: m.CreatedAt, ReadAt: m.ReadAt}
	if !value.Valid() {
		return ports.NotificationRecord{}, ports.ErrInvalidProviderInput
	}
	return value.Clone(), nil
}
