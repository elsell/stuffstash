package config

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var errNotificationPushConfig = errors.New("invalid notification push configuration")

type NotificationPushConfig struct {
	APNSEnabled, FCMEnabled, APNSProduction                                                                 bool
	ServerID, APNSKeyID, APNSTeamID, APNSTopic, APNSKeyFile, FCMProjectID, FCMCredentialsFile, FCMChannelID string
	PageSize                                                                                                int
	PollInterval, PageTimeout, Lease                                                                        time.Duration
	Retry                                                                                                   notification.RetryPolicy
}

func (c NotificationPushConfig) Enabled() bool { return c.APNSEnabled || c.FCMEnabled }
func LoadNotificationPush() (NotificationPushConfig, error) {
	c := NotificationPushConfig{APNSProduction: true, FCMChannelID: "expiration", PageSize: 10, PollInterval: 5 * time.Second, PageTimeout: 30 * time.Second, Lease: time.Minute, Retry: notification.RetryPolicy{MaxAttempts: 6, InitialDelay: time.Minute, MaximumDelay: 15 * time.Minute}}
	const prefix = "STUFF_STASH_PUSH_"
	for _, field := range []struct {
		name  string
		value *bool
	}{{"APNS_ENABLED", &c.APNSEnabled}, {"FCM_ENABLED", &c.FCMEnabled}, {"APNS_PRODUCTION", &c.APNSProduction}} {
		if raw, ok := os.LookupEnv(prefix + field.name); ok {
			value, err := strconv.ParseBool(raw)
			if err != nil {
				return c, errNotificationPushConfig
			}
			*field.value = value
		}
	}
	for _, field := range []struct {
		name  string
		value *string
	}{{"SERVER_ID", &c.ServerID}, {"APNS_KEY_ID", &c.APNSKeyID}, {"APNS_TEAM_ID", &c.APNSTeamID}, {"APNS_TOPIC", &c.APNSTopic}, {"APNS_KEY_FILE", &c.APNSKeyFile}, {"FCM_PROJECT_ID", &c.FCMProjectID}, {"FCM_CREDENTIALS_FILE", &c.FCMCredentialsFile}, {"FCM_CHANNEL_ID", &c.FCMChannelID}} {
		if value, ok := os.LookupEnv(prefix + field.name); ok {
			*field.value = strings.TrimSpace(value)
		}
	}
	for _, field := range []struct {
		name  string
		value *time.Duration
	}{{"POLL_INTERVAL", &c.PollInterval}, {"PAGE_TIMEOUT", &c.PageTimeout}, {"LEASE", &c.Lease}, {"RETRY_INITIAL_DELAY", &c.Retry.InitialDelay}, {"RETRY_MAXIMUM_DELAY", &c.Retry.MaximumDelay}} {
		if raw, ok := os.LookupEnv(prefix + field.name); ok {
			value, err := time.ParseDuration(raw)
			if err != nil {
				return c, errNotificationPushConfig
			}
			*field.value = value
		}
	}
	for _, field := range []struct {
		name  string
		value *int
	}{{"PAGE_SIZE", &c.PageSize}, {"MAX_ATTEMPTS", &c.Retry.MaxAttempts}} {
		if raw, ok := os.LookupEnv(prefix + field.name); ok {
			value, err := strconv.Atoi(raw)
			if err != nil {
				return c, errNotificationPushConfig
			}
			*field.value = value
		}
	}
	if c.PageSize < 1 || c.PageSize > 100 || c.PollInterval < 100*time.Millisecond || c.PageTimeout < time.Second || c.PageTimeout >= c.Lease || c.Lease > time.Hour || c.Retry.Validate() != nil || c.Retry.MaxAttempts > 20 {
		return c, errNotificationPushConfig
	}
	if c.Enabled() {
		server, err := url.Parse(c.ServerID)
		if err != nil || server.Host == "" || (server.Scheme != "http" && server.Scheme != "https") || server.User != nil || server.RawQuery != "" || server.Fragment != "" {
			return c, errNotificationPushConfig
		}
	}
	if c.APNSEnabled && (c.APNSKeyID == "" || c.APNSTeamID == "" || c.APNSTopic == "" || c.APNSKeyFile == "") {
		return c, errNotificationPushConfig
	}
	if c.FCMEnabled && (c.FCMProjectID == "" || c.FCMCredentialsFile == "" || c.FCMChannelID == "") {
		return c, errNotificationPushConfig
	}
	return c, nil
}
