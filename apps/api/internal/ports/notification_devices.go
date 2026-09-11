package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"strings"
	"time"
)

type NotificationDevice struct {
	ID             string
	Scope          NotificationScope
	InstallationID string
	Transport      notification.PushTransport
	Token          notification.DeviceToken
	Active         bool
	Revision       int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (d NotificationDevice) Valid() bool {
	return (d.Transport != notification.PushAPNS || d.Token.Secret() == strings.ToLower(d.Token.Secret())) && d.ID != "" && d.Scope.Valid() && d.InstallationID != "" && len(d.InstallationID) <= 128 && d.Transport.Valid() && !d.Token.Empty() && d.Revision > 0 && !d.CreatedAt.IsZero() && !d.UpdatedAt.Before(d.CreatedAt)
}

type NotificationDeviceRepository interface {
	NotificationDeviceByID(context.Context, NotificationScope, string) (NotificationDevice, bool, error)
	NotificationDeviceByInstallation(context.Context, NotificationScope, string) (NotificationDevice, bool, error)
	ListNotificationDevices(context.Context, NotificationScope, string, int) ([]NotificationDevice, error)
	SaveNotificationDevice(context.Context, NotificationDevice, int64, audit.Record) error
}

// NotificationPushTokenValidator performs local provider-specific token validation.
type NotificationPushTokenValidator interface {
	NormalizeDeviceToken(context.Context, notification.PushTransport, notification.DeviceToken) (notification.DeviceToken, error)
}
