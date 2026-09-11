package push

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var errPushUnavailable = errors.New("notification push provider is unavailable")

type NativeSenders struct{ APNS, FCM ports.NotificationPushSender }

func (s NativeSenders) SendNotification(ctx context.Context, message ports.NotificationPushMessage) (ports.NotificationPushResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, err
	}
	var sender ports.NotificationPushSender
	switch message.Transport {
	case notification.PushAPNS:
		sender = s.APNS
	case notification.PushFCM:
		sender = s.FCM
	}
	if sender == nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errPushUnavailable
	}
	return sender.SendNotification(ctx, message)
}
