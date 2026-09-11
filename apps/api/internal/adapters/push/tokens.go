package push

import (
	"context"
	"encoding/hex"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
)

// NativeTokens validates locally; sending remains a separate provider operation.
type NativeTokens struct{}

func (NativeTokens) NormalizeDeviceToken(ctx context.Context, transport notification.PushTransport, token notification.DeviceToken) (notification.DeviceToken, error) {
	if err := ctx.Err(); err != nil {
		return notification.DeviceToken{}, err
	}
	if token.Empty() || !transport.Valid() {
		return notification.DeviceToken{}, notification.ErrInvalidDeviceToken
	}
	if transport == notification.PushAPNS {
		decoded, err := hex.DecodeString(token.Secret())
		if err != nil {
			return notification.DeviceToken{}, notification.ErrInvalidDeviceToken
		}
		return notification.ParseDeviceToken(hex.EncodeToString(decoded))
	}
	return token, nil
}
