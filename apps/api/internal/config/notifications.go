package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type NotificationConfig struct {
	Enabled      bool
	PageSize     int
	PollInterval time.Duration
	PageTimeout  time.Duration
}

func LoadNotifications() (NotificationConfig, error) {
	cfg := NotificationConfig{Enabled: true, PageSize: 100, PollInterval: 5 * time.Second, PageTimeout: 30 * time.Second}
	const prefix = "STUFF_STASH_NOTIFICATION_"
	if raw, ok := os.LookupEnv(prefix + "WORKER_ENABLED"); ok {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return cfg, fmt.Errorf("invalid notification worker enabled")
		}
		cfg.Enabled = value
	}
	if raw, ok := os.LookupEnv(prefix + "PAGE_SIZE"); ok {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return cfg, fmt.Errorf("invalid notification page size")
		}
		cfg.PageSize = value
	}
	for _, field := range []struct {
		name   string
		target *time.Duration
	}{{"POLL_INTERVAL", &cfg.PollInterval}, {"PAGE_TIMEOUT", &cfg.PageTimeout}} {
		if raw, ok := os.LookupEnv(prefix + field.name); ok {
			value, err := time.ParseDuration(raw)
			if err != nil {
				return cfg, fmt.Errorf("invalid notification %s", field.name)
			}
			*field.target = value
		}
	}
	if cfg.PageSize < 1 || cfg.PageSize > 100 || cfg.PollInterval < 100*time.Millisecond || cfg.PageTimeout < time.Second {
		return cfg, fmt.Errorf("invalid notification worker bounds")
	}
	return cfg, nil
}
