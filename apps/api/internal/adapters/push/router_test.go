package push

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

type nativeSenderFake struct{ calls int }

func (f *nativeSenderFake) SendNotification(context.Context, ports.NotificationPushMessage) (ports.NotificationPushResult, error) {
	f.calls++
	return ports.NotificationPushResult{Outcome: ports.NotificationPushAccepted}, nil
}
func TestNativeSendersNeverCrossTransports(t *testing.T) {
	apns, fcm := &nativeSenderFake{}, &nativeSenderFake{}
	senders := NativeSenders{APNS: apns, FCM: fcm}
	for _, kind := range []notification.PushTransport{notification.PushAPNS, notification.PushFCM} {
		if _, err := senders.SendNotification(context.Background(), ports.NotificationPushMessage{Transport: kind}); err != nil {
			t.Fatal(err)
		}
	}
	if apns.calls != 1 || fcm.calls != 1 {
		t.Fatal("wrong route")
	}
	senders.APNS = nil
	if _, err := senders.SendNotification(context.Background(), ports.NotificationPushMessage{Transport: notification.PushAPNS}); err == nil || fcm.calls != 1 {
		t.Fatal("missing provider fell back")
	}
}
