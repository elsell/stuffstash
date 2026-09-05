package config

import (
	"testing"
	"time"
)

func TestEvaluationConfigurationDefaultsAndOverrides(t *testing.T) {
	defaults, err := (EvaluationConfiguration{}).Settings()
	if err != nil || !defaults.Enabled || defaults.Concurrency != 1 || defaults.DrainLimit != 10 || defaults.Interval != 5*time.Second || defaults.PollInterval != 2*time.Second || defaults.LeaseGrace != 30*time.Second || defaults.MaxAttempts != 2 {
		t.Fatalf("unexpected defaults: %+v %v", defaults, err)
	}
	t.Setenv("STUFF_STASH_EVALUATION_CONCURRENCY", "3")
	t.Setenv("STUFF_STASH_EVALUATION_WORKER_ENABLED", "false")
	t.Setenv("STUFF_STASH_EVALUATION_POLL_INTERVAL", "500ms")
	configured, err := Load().ConversationEvaluations.Settings()
	if err != nil || configured.Enabled || configured.Concurrency != 3 || configured.PollInterval != 500*time.Millisecond {
		t.Fatalf("settings ignored: %+v %v", configured, err)
	}
}
func TestEvaluationConfigurationRejectsMalformedAndUnsafeValues(t *testing.T) {
	for key, value := range map[string]string{
		"WORKER_ENABLED": "perhaps", "DRAIN_LIMIT": "101", "CONCURRENCY": "9", "INTERVAL": "0s", "POLL_INTERVAL": "31s", "LEASE_GRACE": "-1s", "MAX_ATTEMPTS": "11",
	} {
		t.Run(key, func(t *testing.T) {
			t.Setenv("STUFF_STASH_EVALUATION_"+key, value)
			if _, err := Load().ConversationEvaluations.Settings(); err == nil {
				t.Fatal("invalid setting accepted")
			}
		})
	}
}
