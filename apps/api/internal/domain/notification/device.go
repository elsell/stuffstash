package notification

import "errors"

type PushTransport string

const (
	PushAPNS PushTransport = "apns"
	PushFCM  PushTransport = "fcm"
)

var ErrInvalidDeviceToken = errors.New("invalid notification device token")

// DeviceToken deliberately redacts ordinary formatting. Only infrastructure
// adapters should access Secret for persistence or provider delivery.
type DeviceToken struct{ value string }

func ParseDeviceToken(value string) (DeviceToken, error) {
	if len(value) == 0 || len(value) > 4096 {
		return DeviceToken{}, ErrInvalidDeviceToken
	}
	for _, r := range value {
		if r < 33 || r > 126 {
			return DeviceToken{}, ErrInvalidDeviceToken
		}
	}
	return DeviceToken{value: value}, nil
}
func (t DeviceToken) Secret() string               { return t.value }
func (t DeviceToken) Empty() bool                  { return t.value == "" }
func (t DeviceToken) String() string               { return "[redacted]" }
func (t DeviceToken) GoString() string             { return "[redacted]" }
func (t DeviceToken) MarshalJSON() ([]byte, error) { return []byte(`"[redacted]"`), nil }
func (t PushTransport) Valid() bool                { return t == PushAPNS || t == PushFCM }
