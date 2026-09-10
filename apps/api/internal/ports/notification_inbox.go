package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"time"
)

type NotificationRecord struct {
	ID        string
	Scope     NotificationScope
	Milestone notification.Milestone
	CreatedAt time.Time
	ReadAt    *time.Time
}

func (r NotificationRecord) Clone() NotificationRecord {
	if r.ReadAt != nil {
		value := *r.ReadAt
		r.ReadAt = &value
	}
	return r
}

func (r NotificationRecord) Valid() bool {
	return r.ID != "" && r.Scope.Valid() && r.Milestone.AssetID != "" && r.Milestone.Date.Value() != "" &&
		(r.Milestone.Kind == notification.MilestoneUpcoming || r.Milestone.Kind == notification.MilestoneExpired) && !r.CreatedAt.IsZero()
}

type NotificationInboxRepository interface {
	InsertNotification(ctx context.Context, value NotificationRecord, record audit.Record) (NotificationRecord, bool, error)
	NotificationByID(ctx context.Context, scope NotificationScope, id string) (NotificationRecord, bool, error)
	ListNotifications(ctx context.Context, scope NotificationScope, beforeID string, limit int) ([]NotificationRecord, error)
	MarkNotificationRead(ctx context.Context, scope NotificationScope, id string, at time.Time, record audit.Record) (bool, error)
}
