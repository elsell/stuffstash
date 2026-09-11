package config

import "testing"

func TestNotificationWorkerConfiguration(t *testing.T) {
	cfg, err := LoadNotifications()
	if err != nil || !cfg.Enabled || cfg.PageSize != 100 {
		t.Fatalf("defaults %+v %v", cfg, err)
	}
	for key, value := range map[string]string{"STUFF_STASH_NOTIFICATION_WORKER_ENABLED": "bad", "STUFF_STASH_NOTIFICATION_POLL_INTERVAL": "0s", "STUFF_STASH_NOTIFICATION_PAGE_TIMEOUT": "0s", "STUFF_STASH_NOTIFICATION_PAGE_SIZE": "101"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, value)
			if _, err := LoadNotifications(); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
	t.Setenv("STUFF_STASH_NOTIFICATION_WORKER_ENABLED", "false")
	cfg, err = LoadNotifications()
	if err != nil || cfg.Enabled {
		t.Fatal("cannot disable worker")
	}
}
