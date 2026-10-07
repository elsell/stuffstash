package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"unicode/utf8"
)

func isNotificationCommand(o Options) bool {
	return len(o.Command) > 0 && o.Command[0] == "notifications"
}
func validateNotifications(o Options, scope bool) error {
	valid := false
	if len(o.Command) == 2 {
		switch o.Command[1] {
		case "list", "unread-count", "read-all":
			valid = true
		}
	}
	if len(o.Command) == 3 && o.Command[2] != "" {
		switch o.Command[1] {
		case "show", "read", "unread":
			valid = true
		}
	}
	if !valid {
		return ports.Failure("usage", "Use notifications list, show ID, unread-count, read ID, unread ID, or read-all.")
	}
	if o.Page.Cursor != "" && o.Command[1] != "list" && o.Command[1] != "unread-count" && o.Command[1] != "read-all" {
		return ports.Failure("usage", "Use --cursor with notifications list, unread-count, or read-all.")
	}
	if o.UnreadOnly && o.Command[1] != "list" {
		return ports.Failure("usage", "Use --unread-only with notifications list.")
	}
	if utf8.RuneCountInString(o.Page.Cursor) > 128 || o.Command[1] == "list" && (o.Page.Limit < 1 || o.Page.Limit > 100) {
		return ports.Failure("usage", "Use a limit from 1 to 100 and a cursor with at most 128 characters.")
	}
	if o.IdempotencyKey != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" {
		return ports.Failure("usage", "Notification commands do not accept retry keys or asset fields. Remove those options.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved context.")
	}
	return nil
}
func (r Runner) notificationsCommand(ctx context.Context, o Options, token string) error {
	if r.NotificationsAPI == nil {
		return ports.Failure("configuration", "Notification commands are not available. Update the CLI and try again.")
	}
	api, err := r.NotificationsAPI(o.Server, token)
	if err != nil {
		return err
	}
	action := o.Command[1]
	if action == "read" || action == "unread" || action == "read-all" {
		if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory)); err != nil {
			return err
		}
	}
	var result any
	switch action {
	case "list":
		result, err = api.Notifications(ctx, o.Scope, o.Page, o.UnreadOnly)
	case "show":
		result, err = api.Notification(ctx, o.Scope, o.Command[2])
	case "unread-count":
		result, err = api.NotificationUnreadCount(ctx, o.Scope, o.Page.Cursor)
	case "read", "unread":
		result, err = api.SetNotificationRead(ctx, o.Scope, o.Command[2], action == "read")
	case "read-all":
		result, err = api.ReadAllNotifications(ctx, o.Scope, o.Page.Cursor)
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.notification.command.completed")
	return r.Output.Result(result)
}
