package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type EvaluationSettings struct {
	Enabled                              bool
	DrainLimit, Concurrency, MaxAttempts int
	Interval, PollInterval, LeaseGrace   time.Duration
}
type EvaluationConfiguration struct {
	settings EvaluationSettings
	captured bool
	err      error
}

func defaultEvaluationSettings() EvaluationSettings {
	return EvaluationSettings{Enabled: true, DrainLimit: 10, Concurrency: 1, MaxAttempts: 2, Interval: 5 * time.Second, PollInterval: 2 * time.Second, LeaseGrace: 30 * time.Second}
}
func (c EvaluationConfiguration) Settings() (EvaluationSettings, error) {
	if c.err != nil {
		return EvaluationSettings{}, c.err
	}
	if !c.captured {
		return defaultEvaluationSettings(), nil
	}
	return c.settings, nil
}
func loadEvaluationConfiguration() EvaluationConfiguration {
	c := EvaluationConfiguration{settings: defaultEvaluationSettings(), captured: true}
	const prefix = "STUFF_STASH_EVALUATION_"
	if raw := strings.TrimSpace(os.Getenv(prefix + "WORKER_ENABLED")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			c.err = fmt.Errorf("%sWORKER_ENABLED must be boolean", prefix)
			return c
		}
		c.settings.Enabled = value
	}
	for _, entry := range []struct {
		name   string
		target *int
		max    int
	}{
		{"DRAIN_LIMIT", &c.settings.DrainLimit, ports.MaxEvaluationRunPageLimit},
		{"CONCURRENCY", &c.settings.Concurrency, ports.MaxEvaluationWorkerConcurrency},
		{"MAX_ATTEMPTS", &c.settings.MaxAttempts, agentmodel.MaxEvaluationRunAttempts},
	} {
		raw := strings.TrimSpace(os.Getenv(prefix + entry.name))
		if raw == "" {
			continue
		}
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > entry.max {
			c.err = fmt.Errorf("%s%s must be between 1 and %d", prefix, entry.name, entry.max)
			return c
		}
		*entry.target = value
	}
	for _, entry := range []struct {
		name     string
		target   *time.Duration
		min, max time.Duration
	}{
		{"INTERVAL", &c.settings.Interval, 100 * time.Millisecond, time.Hour},
		{"POLL_INTERVAL", &c.settings.PollInterval, 100 * time.Millisecond, 30 * time.Second},
		{"LEASE_GRACE", &c.settings.LeaseGrace, time.Second, 5 * time.Minute},
	} {
		raw := strings.TrimSpace(os.Getenv(prefix + entry.name))
		if raw == "" {
			continue
		}
		value, err := time.ParseDuration(raw)
		if err != nil || value < entry.min || value > entry.max {
			c.err = fmt.Errorf("%s%s must be between %s and %s", prefix, entry.name, entry.min, entry.max)
			return c
		}
		*entry.target = value
	}
	return c
}
