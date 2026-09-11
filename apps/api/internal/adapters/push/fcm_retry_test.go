package push

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
	"testing"
	"time"
)

type fcmRetryTransport struct {
	apnsTransportFake
	retry string
}

func (f *fcmRetryTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := f.apnsTransportFake.RoundTrip(request)
	if err == nil {
		response.Header.Set("Retry-After", f.retry)
	}
	return response, err
}
func TestFCMRetryDeadlineSurvivesMalformedResponse(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		header string
		status int
		delay  time.Duration
	}{{"120", 503, 2 * time.Minute}, {now.Add(time.Hour).Format(http.TimeFormat), 503, time.Hour}, {"", 429, time.Minute}, {"2", 429, time.Minute}, {"bad", 503, 0}, {"-1", 503, 0}, {"99999999999999999999999", 503, 0}} {
		transport := &fcmRetryTransport{apnsTransportFake: apnsTransportFake{status: tc.status, body: "invalid provider response"}, retry: tc.header}
		sender, err := NewFCM(FCMConfig{ProjectID: "test-project", ServerID: "https://api.test", ChannelID: "expiration"}, fcmAuthFake{}, transport, &authClock{now: now})
		if err != nil {
			t.Fatal(err)
		}
		token, _ := notification.ParseDeviceToken("token")
		result, err := sender.SendNotification(context.Background(), ports.NotificationPushMessage{DeliveryID: "delivery", NotificationID: "notice", Scope: ports.NotificationScope{TenantID: "tenant", InventoryID: "inventory", PrincipalID: "owner"}, Transport: notification.PushFCM, Token: token})
		if err != nil || result.Outcome != ports.NotificationPushRetry {
			t.Fatal("not retryable")
		}
		want := time.Time{}
		if tc.delay > 0 {
			want = now.Add(tc.delay)
		}
		if !result.RetryNotBefore.Equal(want) {
			t.Fatalf("header %s: got %v want %v", tc.header, result.RetryNotBefore, want)
		}
	}
}
