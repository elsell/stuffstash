package push

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type authTokenFake struct{}

func (authTokenFake) Token() (string, error) { return "provider-token", nil }

type apnsTransportFake struct {
	status   int
	body     string
	requests int
	request  *http.Request
	payload  map[string]any
}

func (f *apnsTransportFake) RoundTrip(r *http.Request) (*http.Response, error) {
	f.requests++
	f.request = r
	_ = json.NewDecoder(r.Body).Decode(&f.payload)
	return &http.Response{StatusCode: f.status, Body: io.NopCloser(strings.NewReader(f.body)), Header: http.Header{"Location": []string{"https://outside.test"}}}, nil
}
func TestAPNSRequestAndResponseClassification(t *testing.T) {
	for _, scenario := range []struct {
		status  int
		body    string
		outcome ports.NotificationPushOutcome
	}{{200, "", ports.NotificationPushAccepted}, {400, `{"reason":"BadDeviceToken"}`, ports.NotificationPushInvalidDevice}, {410, `{"reason":"Unregistered"}`, ports.NotificationPushRetry}, {410, `{"reason":"Unregistered","timestamp":1788998400000}`, ports.NotificationPushInvalidDevice}, {403, `{"reason":"InvalidProviderToken"}`, ports.NotificationPushRetry}, {400, `{"reason":"DeviceTokenNotForTopic"}`, ports.NotificationPushRetry}, {500, "private failure details", ports.NotificationPushRetry}, {302, "", ports.NotificationPushRetry}} {
		transport := &apnsTransportFake{status: scenario.status, body: scenario.body}
		sender, err := NewAPNS(APNSConfig{Production: true, Topic: "org.stuffstash.mobile", ServerID: "https://api.test"}, authTokenFake{}, transport)
		if err != nil {
			t.Fatal(err)
		}
		token, _ := notification.ParseDeviceToken("abcd")
		outcome, err := sender.SendNotification(context.Background(), ports.NotificationPushMessage{DeliveryID: "delivery", NotificationID: "notice", Scope: ports.NotificationScope{TenantID: "tenant", InventoryID: "inventory", PrincipalID: "owner"}, Transport: notification.PushAPNS, Token: token, Title: "Stuff Stash", Body: "An item is expiring soon."})
		if err != nil || outcome.Outcome != scenario.outcome {
			t.Fatalf("status %d: %s %v", scenario.status, outcome, err)
		}
		if scenario.status == 410 && outcome.Outcome == ports.NotificationPushInvalidDevice && !outcome.InvalidatedAt.Equal(time.UnixMilli(1788998400000)) {
			t.Fatal("invalidation timestamp lost")
		}
		if transport.requests != 1 || transport.request.URL.String() != "https://api.push.apple.com/3/device/abcd" || transport.request.Header.Get("Authorization") != "bearer provider-token" || transport.request.Header.Get("apns-topic") != "org.stuffstash.mobile" || transport.request.Header.Get("apns-expiration") != "0" {
			t.Fatal("unsafe APNs request")
		}
		if transport.payload["notificationId"] != "notice" || transport.payload["serverId"] != "https://api.test" {
			t.Fatal("missing scoped navigation data")
		}
	}
}

type blockingAPNSTransport struct{ started chan struct{} }

func (f blockingAPNSTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	close(f.started)
	<-request.Context().Done()
	return nil, request.Context().Err()
}
func TestAPNSHonorsInflightCancellation(t *testing.T) {
	transport := blockingAPNSTransport{started: make(chan struct{})}
	sender, err := NewAPNS(APNSConfig{Topic: "org.stuffstash.mobile", ServerID: "https://api.test"}, authTokenFake{}, transport)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := notification.ParseDeviceToken("abcd")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := sender.SendNotification(ctx, ports.NotificationPushMessage{DeliveryID: "delivery", NotificationID: "notice", Scope: ports.NotificationScope{TenantID: "tenant", InventoryID: "inventory", PrincipalID: "owner"}, Transport: notification.PushAPNS, Token: token})
		done <- err
	}()
	<-transport.started
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("cancellation lost: %v", err)
	}
}
func TestAPNSBoundsProviderResponses(t *testing.T) {
	transport := &apnsTransportFake{status: 410, body: `{"reason":"Unregistered","padding":"` + strings.Repeat("x", 4096) + `"}`}
	sender, err := NewAPNS(APNSConfig{Topic: "org.stuffstash.mobile", ServerID: "https://api.test"}, authTokenFake{}, transport)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := notification.ParseDeviceToken("abcd")
	outcome, err := sender.SendNotification(context.Background(), ports.NotificationPushMessage{DeliveryID: "delivery", NotificationID: "notice", Scope: ports.NotificationScope{TenantID: "tenant", InventoryID: "inventory", PrincipalID: "owner"}, Transport: notification.PushAPNS, Token: token})
	if err != nil || outcome.Outcome != ports.NotificationPushRetry {
		t.Fatal("oversized response revoked device")
	}
}
