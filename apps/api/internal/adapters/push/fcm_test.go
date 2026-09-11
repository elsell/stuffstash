package push

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
	"testing"
	"time"
)

type fcmAuthFake struct{}

func (fcmAuthFake) Token(context.Context) (string, error) { return "access-token", nil }
func TestFCMRequestAndProviderClassification(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   ports.NotificationPushOutcome
	}{
		{200, `{"name":"projects/test-project/messages/message-id"}`, ports.NotificationPushAccepted},
		{200, `{}`, ports.NotificationPushRetry},
		{404, `{"error":{"details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`, ports.NotificationPushInvalidDevice},
		{404, `{"error":{"status":"NOT_FOUND"}}`, ports.NotificationPushRetry},
		{400, `{"error":{"status":"INVALID_ARGUMENT"}}`, ports.NotificationPushRetry},
		{403, `{"error":{"status":"SENDER_ID_MISMATCH"}}`, ports.NotificationPushRetry},
		{302, ``, ports.NotificationPushRetry}, {503, `private provider data`, ports.NotificationPushRetry},
	} {
		transport := &apnsTransportFake{status: tc.status, body: tc.body}
		sender, err := NewFCM(FCMConfig{ProjectID: "test-project", ServerID: "https://api.test", ChannelID: "expiration"}, fcmAuthFake{}, transport, &authClock{now: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			t.Fatal(err)
		}
		token, _ := notification.ParseDeviceToken("device-token")
		result, err := sender.SendNotification(context.Background(), ports.NotificationPushMessage{DeliveryID: "delivery", NotificationID: "notice", Scope: ports.NotificationScope{TenantID: "tenant", InventoryID: "inventory", PrincipalID: "owner"}, Transport: notification.PushFCM, Token: token, Title: "Stuff Stash", Body: "An item is expiring soon."})
		if err != nil || result.Outcome != tc.want {
			t.Fatalf("status%d: %+v %v", tc.status, result, err)
		}
		if transport.requests != 1 || transport.request.URL.String() != "https://fcm.googleapis.com/v1/projects/test-project/messages:send" || transport.request.Header.Get("Authorization") != "Bearer access-token" {
			t.Fatal("unsafe FCM request")
		}
		message := transport.payload["message"].(map[string]any)
		data := message["data"].(map[string]any)
		if message["token"] != "device-token" || data["notificationId"] != "notice" || data["serverId"] != "https://api.test" {
			t.Fatal("scoped message missing")
		}
		android := message["android"].(map[string]any)
		if android["priority"] != "HIGH" || android["ttl"] != "0s" || android["notification"].(map[string]any)["channel_id"] != "expiration" {
			t.Fatal("android config missing")
		}
	}
}

type failingFCMAuth struct{}

func (failingFCMAuth) Token(context.Context) (string, error) {
	return "", errors.New("private-key-material")
}
func TestFCMDoesNotLeakCredentialsAndHonorsCancellation(t *testing.T) {
	transport := &apnsTransportFake{status: 200, body: `{"name":"projects/test-project/messages/id"}`}
	sender, err := NewFCM(FCMConfig{ProjectID: "test-project", ServerID: "https://api.test", ChannelID: "expiration"}, failingFCMAuth{}, transport, &authClock{now: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	token, _ := notification.ParseDeviceToken("device-token")
	message := ports.NotificationPushMessage{DeliveryID: "delivery", NotificationID: "notice", Scope: ports.NotificationScope{TenantID: "tenant", InventoryID: "inventory", PrincipalID: "owner"}, Transport: notification.PushFCM, Token: token}
	_, err = sender.SendNotification(context.Background(), message)
	if err == nil || strings.Contains(err.Error(), "private-key") || transport.requests != 0 {
		t.Fatal("credential error leaked or request sent")
	}
	blocked := blockingAPNSTransport{started: make(chan struct{})}
	sender, err = NewFCM(FCMConfig{ProjectID: "test-project", ServerID: "https://api.test", ChannelID: "expiration"}, fcmAuthFake{}, blocked, &authClock{now: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := sender.SendNotification(ctx, message); done <- err }()
	<-blocked.started
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("lost cancellation: %v", err)
	}
}
func TestFCMRejectsInvalidConfiguration(t *testing.T) {
	for _, project := range []string{"../other", "project/other", "", "PROJECT"} {
		if _, err := NewFCM(FCMConfig{ProjectID: project, ServerID: "https://api.test", ChannelID: "expiration"}, fcmAuthFake{}, nil, &authClock{now: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)}); err == nil {
			t.Fatal("invalid project accepted")
		}
	}
}
