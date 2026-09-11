package app

import "github.com/stuffstash/stuff-stash/internal/app/notifications"

func (a App) Notifications() notifications.Service { return a.notificationService }
