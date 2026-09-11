package bootstrap

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/push"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
	"os"
)

var errPushCredentials = errors.New("notification push credentials are unavailable or invalid")

func buildNotificationPush(cfg config.NotificationPushConfig, clock ports.Clock) (ports.NotificationPushSender, error) {
	if !cfg.Enabled() {
		return nil, nil
	}
	senders := push.NativeSenders{}
	if cfg.APNSEnabled {
		data, err := readPushCredential(cfg.APNSKeyFile)
		if err != nil {
			return nil, err
		}
		auth, err := push.NewAPNSAuth(cfg.APNSKeyID, cfg.APNSTeamID, data, clock)
		if err != nil {
			return nil, errPushCredentials
		}
		senders.APNS, err = push.NewAPNS(push.APNSConfig{Production: cfg.APNSProduction, Topic: cfg.APNSTopic, ServerID: cfg.ServerID}, auth, nil)
		if err != nil {
			return nil, errPushCredentials
		}
	}
	if cfg.FCMEnabled {
		data, err := readPushCredential(cfg.FCMCredentialsFile)
		if err != nil {
			return nil, err
		}
		auth, err := push.NewFCMAuth(data, clock, nil)
		if err != nil {
			return nil, errPushCredentials
		}
		senders.FCM, err = push.NewFCM(push.FCMConfig{ProjectID: cfg.FCMProjectID, ServerID: cfg.ServerID, ChannelID: cfg.FCMChannelID}, auth, nil, clock)
		if err != nil {
			return nil, errPushCredentials
		}
	}
	return senders, nil
}
func readPushCredential(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errPushCredentials
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 65537))
	if err != nil || len(data) == 0 || len(data) > 65536 {
		return nil, errPushCredentials
	}
	return data, nil
}
