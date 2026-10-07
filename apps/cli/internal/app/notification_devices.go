package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isNotificationDeviceCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "notification-devices"
}
func validateNotificationDevices(o Options, scope bool) error {
	if !(len(o.Command) == 2 && isDeviceRegistration(o)) && (len(o.Command) != 3 || o.Command[2] == "" || (o.Command[1] != "show" && o.Command[1] != "remove")) {
		return ports.Failure("usage", "Use notification-devices register, show INSTALLATION_ID, or remove DEVICE_ID --revision N.")
	}
	if o.Command[1] != "remove" && o.Revision != -1 {
		return ports.Failure("usage", "Use --revision with notification-devices remove.")
	}
	if o.Revision != -1 && o.Revision < 1 {
		return ports.Failure("usage", "The revision must be positive. Use the device revision from notification-devices show.")
	}
	if o.IdempotencyKey != "" || o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.Page.Cursor != "" {
		return ports.Failure("usage", "Device commands do not accept asset fields, cursors, or retry keys. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved context.")
	}
	return nil
}
func (r Runner) notificationDevicesCommand(ctx context.Context, o Options, token string) error {
	if r.NotificationDevicesAPI == nil {
		return ports.Failure("configuration", "Device commands are not available. Update the CLI and try again.")
	}
	api, err := r.NotificationDevicesAPI(o.Server, token)
	if err != nil {
		return err
	}
	if isDeviceRegistration(o) {
		return r.registerNotificationDevice(ctx, o, api)
	}
	if o.Command[1] == "show" {
		result, err := api.NotificationDevice(ctx, o.Scope, o.Command[2])
		if err != nil {
			return err
		}
		return r.Output.Result(result)
	}
	revision := o.Revision
	if revision == -1 {
		if r.TextInput == nil || o.NoInput || o.JSON {
			return ports.Failure("usage", "Supply --revision N. Use the device revision from notification-devices show INSTALLATION_ID.")
		}
		text, err := r.TextInput.ReadText(ctx, "Device revision", 19)
		if err != nil {
			return err
		}
		revision, err = strconv.ParseInt(text, 10, 64)
		if err != nil || revision < 1 {
			return ports.Failure("usage", "The revision must be a positive integer. Examine the device revision and try again.")
		}
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Device: " + strconv.Quote(o.Command[2]) + ". Revision: " + strconv.FormatInt(revision, 10)); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Remove notification device", "Remove", "Stop notifications for this device registration."); err != nil {
		return err
	}
	result, err := api.RemoveNotificationDevice(ctx, o.Scope, o.Command[2], revision)
	if err != nil {
		var failure *ports.Error
		if errors.As(err, &failure) && failure.Category == "conflict" {
			return ports.Failure("conflict", "The device registration changed. Run notification-devices show INSTALLATION_ID and review the current revision before you try again.")
		}
		return err
	}
	r.Observer.Event(ctx, "cli.notification.device.removed")
	return r.Output.Result(result)
}
