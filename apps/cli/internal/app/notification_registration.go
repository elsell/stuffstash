package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isDeviceRegistration(o Options) bool {
	return isNotificationDeviceCommand(o) && len(o.Command) > 1 && o.Command[1] == "register"
}
func (r Runner) prepareDeviceRegistration(ctx context.Context, o Options) (Options, error) {
	if r.TextInput == nil || r.Picker == nil || r.SecretInput == nil || o.NoInput || o.JSON {
		return o, ports.Failure("usage", "Supply --input FILE with installationId, transport, token, and revision to register a device.")
	}
	installation, err := r.TextInput.ReadText(ctx, "Installation ID", 128)
	if err != nil {
		return o, err
	}
	transport, err := r.Picker.Pick(ctx, "Push service", []ports.Choice{{ID: "apns", Label: "Apple Push Notification service"}, {ID: "fcm", Label: "Firebase Cloud Messaging"}})
	if err != nil {
		return o, err
	}
	if transport != "apns" && transport != "fcm" {
		return o, context.Canceled
	}
	revisionText, err := r.TextInput.ReadText(ctx, "Device revision (0 for first registration)", 19)
	if err != nil {
		return o, err
	}
	revision, err := strconv.ParseInt(revisionText, 10, 64)
	if err != nil || revision < 0 {
		return o, ports.Failure("usage", "The device revision must be zero or a positive integer.")
	}
	secret, err := r.SecretInput.ReadSecret(ctx, "Device token", 4095)
	if err != nil {
		return o, err
	}
	o.RequestBody, err = json.Marshal(struct {
		InstallationID string `json:"installationId"`
		Transport      string `json:"transport"`
		Token          string `json:"token"`
		Revision       int64  `json:"revision"`
	}{installation, transport, secret, revision})
	return o, err
}
func (r Runner) registerNotificationDevice(ctx context.Context, o Options, api ports.NotificationDevicesAPI) error {
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory)); err != nil {
		return err
	}
	result, err := api.RegisterNotificationDevice(ctx, o.Scope, o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The registration result is unknown. Run notification-devices show INSTALLATION_ID before you retry.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.notification.device.registered")
	return r.Output.Result(result)
}
