package bootstrap

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type printSettings struct {
	directory                                                     string
	heartbeat, poll, readiness, observe, safety, minimum, maximum time.Duration
	artifactBytes                                                 int64
}

func printConfig(options app.Options, getenv func(string) string) (printSettings, error) {
	c := printSettings{directory: options.JournalDirectory, artifactBytes: 16 << 20}
	durations := []struct {
		name              string
		target            *time.Duration
		fallback, maximum time.Duration
	}{
		{"HEARTBEAT_INTERVAL", &c.heartbeat, 5 * time.Second, time.Minute},
		{"POLL_INTERVAL", &c.poll, 2 * time.Second, time.Minute},
		{"READINESS_TIMEOUT", &c.readiness, 5 * time.Second, time.Minute},
		{"OBSERVE_INTERVAL", &c.observe, time.Second, 10 * time.Second},
		{"LEASE_SAFETY", &c.safety, 5 * time.Second, 30 * time.Second},
		{"BACKOFF_MIN", &c.minimum, time.Second, time.Minute},
		{"BACKOFF_MAX", &c.maximum, 30 * time.Second, 5 * time.Minute},
	}
	for _, setting := range durations {
		value := setting.fallback
		if raw := getenv("STUFF_STASH_CLI_PRINT_" + setting.name); raw != "" {
			var err error
			value, err = time.ParseDuration(raw)
			if err != nil || value < 100*time.Millisecond || value > setting.maximum {
				return c, ports.Failure("configuration", "STUFF_STASH_CLI_PRINT_"+setting.name+" is invalid. Use a duration from 100ms to "+setting.maximum.String()+", such as 1s.")
			}
		}
		*setting.target = value
	}
	if c.minimum > c.maximum {
		return c, ports.Failure("configuration", "STUFF_STASH_CLI_PRINT_BACKOFF_MIN exceeds STUFF_STASH_CLI_PRINT_BACKOFF_MAX. Set the minimum to a duration less than or equal to the maximum.")
	}
	if raw := getenv("STUFF_STASH_CLI_PRINT_MAX_ARTIFACT_BYTES"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 || value > 64<<20 {
			return c, ports.Failure("configuration", "Set STUFF_STASH_CLI_PRINT_MAX_ARTIFACT_BYTES to an integer from 1 through 67108864. The unit is bytes.")
		}
		c.artifactBytes = value
	}
	if c.directory == "" {
		directory, err := os.UserConfigDir()
		if err != nil {
			return c, ports.Failure("configuration", "Set STUFF_STASH_CLI_PRINT_STATE_DIRECTORY to a directory on persistent storage.")
		}
		c.directory = filepath.Join(directory, "stuffstash", "print-state")
	}
	return c, nil
}
