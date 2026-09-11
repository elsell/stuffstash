package push

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var errAPNSRequest = errors.New("APNs delivery is unavailable")
var apnsTopic = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,254}$`)

type APNSConfig struct {
	Production bool
	Topic      string
	ServerID   string
}
type apnsTokenSource interface{ Token() (string, error) }
type APNS struct {
	config   APNSConfig
	auth     apnsTokenSource
	client   *http.Client
	endpoint string
}

func NewAPNS(config APNSConfig, auth apnsTokenSource, transport http.RoundTripper) (*APNS, error) {
	server, err := url.Parse(config.ServerID)
	if err != nil || server.Host == "" || (server.Scheme != "https" && server.Scheme != "http") || server.User != nil || server.RawQuery != "" || server.Fragment != "" || !apnsTopic.MatchString(config.Topic) || auth == nil {
		return nil, errAPNSCredentials
	}
	if transport == nil {
		base := http.DefaultTransport.(*http.Transport).Clone()
		base.ForceAttemptHTTP2 = true
		transport = base
	}
	endpoint := "https://api.sandbox.push.apple.com"
	if config.Production {
		endpoint = "https://api.push.apple.com"
	}
	return &APNS{config: config, auth: auth, endpoint: endpoint, client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (a *APNS) SendNotification(ctx context.Context, message ports.NotificationPushMessage) (ports.NotificationPushResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, err
	}
	token := message.Token.Secret()
	if message.Transport != notification.PushAPNS || token == "" || len(token) > 4096 || message.DeliveryID == "" || message.NotificationID == "" || !message.Scope.Valid() {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errAPNSRequest
	}
	if _, err := hex.DecodeString(token); err != nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errAPNSRequest
	}
	payload, err := json.Marshal(map[string]any{"aps": map[string]any{"alert": map[string]string{"title": message.Title, "body": message.Body}, "sound": "default"}, "notificationId": message.NotificationID, "tenantId": message.Scope.TenantID.String(), "inventoryId": message.Scope.InventoryID.String(), "principalId": message.Scope.PrincipalID.String(), "serverId": a.config.ServerID})
	if err != nil || len(payload) > 4096 {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errAPNSRequest
	}
	auth, err := a.auth.Token()
	if err != nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errAPNSAuthentication
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint+"/3/device/"+strings.ToLower(token), bytes.NewReader(payload))
	if err != nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errAPNSRequest
	}
	collapse := sha256.Sum256([]byte(message.DeliveryID))
	request.Header.Set("Authorization", "bearer "+auth)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("apns-topic", a.config.Topic)
	request.Header.Set("apns-push-type", "alert")
	request.Header.Set("apns-priority", "10")
	request.Header.Set("apns-expiration", "0")
	request.Header.Set("apns-collapse-id", hex.EncodeToString(collapse[:]))
	response, err := a.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, ctx.Err()
		}
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, errAPNSRequest
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushAccepted}, nil
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil || len(body) > 4096 {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, nil
	}
	var failure struct {
		Reason    string `json:"reason"`
		Timestamp int64  `json:"timestamp"`
	}
	if json.Unmarshal(body, &failure) != nil {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, nil
	}
	if response.StatusCode == http.StatusBadRequest && failure.Reason == "BadDeviceToken" {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushInvalidDevice}, nil
	}
	if response.StatusCode == http.StatusGone && failure.Reason == "Unregistered" && failure.Timestamp > 0 && failure.Timestamp <= 253402300799999 {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushInvalidDevice, InvalidatedAt: time.UnixMilli(failure.Timestamp).UTC()}, nil
	}
	return ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}, nil
}
