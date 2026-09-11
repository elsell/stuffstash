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
)

var errFCM = errors.New("FCM delivery is unavailable")
var fcmProject = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

type FCMConfig struct{ ProjectID, ServerID, ChannelID string }
type fcmTokenSource interface {
	Token(context.Context) (string, error)
}
type FCM struct {
	config FCMConfig
	auth   fcmTokenSource
	client *http.Client
}

func NewFCM(config FCMConfig, auth fcmTokenSource, transport http.RoundTripper) (*FCM, error) {
	server, err := url.Parse(config.ServerID)
	if err != nil || server.Host == "" || (server.Scheme != "https" && server.Scheme != "http") || server.User != nil || server.RawQuery != "" || server.Fragment != "" || !fcmProject.MatchString(config.ProjectID) || strings.TrimSpace(config.ChannelID) == "" || len(config.ChannelID) > 100 || auth == nil {
		return nil, errFCM
	}
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	return &FCM{config: config, auth: auth, client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (f *FCM) SendNotification(ctx context.Context, message ports.NotificationPushMessage) (ports.NotificationPushResult, error) {
	retry := ports.NotificationPushResult{Outcome: ports.NotificationPushRetry}
	if ctx.Err() != nil {
		return retry, ctx.Err()
	}
	if message.Transport != notification.PushFCM || message.Token.Empty() || message.DeliveryID == "" || message.NotificationID == "" || !message.Scope.Valid() {
		return retry, errFCM
	}
	tag := sha256.Sum256([]byte(message.DeliveryID))
	payload, err := json.Marshal(map[string]any{"message": map[string]any{"token": message.Token.Secret(), "notification": map[string]string{"title": message.Title, "body": message.Body}, "data": map[string]string{"notificationId": message.NotificationID, "tenantId": message.Scope.TenantID.String(), "inventoryId": message.Scope.InventoryID.String(), "principalId": message.Scope.PrincipalID.String(), "serverId": f.config.ServerID}, "android": map[string]any{"ttl": "0s", "priority": "HIGH", "notification": map[string]string{"channel_id": f.config.ChannelID, "tag": hex.EncodeToString(tag[:])}}}})
	if err != nil || len(payload) > 4096 {
		return retry, errFCM
	}
	token, err := f.auth.Token(ctx)
	if err != nil || strings.TrimSpace(token) == "" {
		if ctx.Err() != nil {
			return retry, ctx.Err()
		}
		return retry, errFCM
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/v1/projects/"+f.config.ProjectID+"/messages:send", bytes.NewReader(payload))
	if err != nil {
		return retry, errFCM
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := f.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return retry, ctx.Err()
		}
		return retry, errFCM
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil || len(body) > 4096 {
		return retry, nil
	}
	var output struct {
		Name  string `json:"name"`
		Error struct {
			Details []struct {
				Type string `json:"@type"`
				Code string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &output) != nil {
		return retry, nil
	}
	prefix := "projects/" + f.config.ProjectID + "/messages/"
	if response.StatusCode == http.StatusOK && strings.HasPrefix(output.Name, prefix) && len(output.Name) > len(prefix) {
		return ports.NotificationPushResult{Outcome: ports.NotificationPushAccepted}, nil
	}
	if response.StatusCode == http.StatusNotFound {
		for _, detail := range output.Error.Details {
			if detail.Type == "type.googleapis.com/google.firebase.fcm.v1.FcmError" && detail.Code == "UNREGISTERED" {
				return ports.NotificationPushResult{Outcome: ports.NotificationPushInvalidDevice}, nil
			}
		}
	}
	return retry, nil
}
