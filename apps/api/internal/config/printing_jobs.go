package config

import (
	"errors"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"strconv"
	"time"
)

type PrintJobConfig struct{ MaxCopies, MaxArtifactBytes, ArtifactTTL, TerminalTTL, Lease, ReadinessMaxAge, CleanupInterval string }

func loadPrintJobs() PrintJobConfig {
	return PrintJobConfig{MaxCopies: envOrDefault("STUFF_STASH_PRINT_MAX_COPIES", "20"), MaxArtifactBytes: envOrDefault("STUFF_STASH_PRINT_ARTIFACT_MAX_BYTES", "1048576"), ArtifactTTL: envOrDefault("STUFF_STASH_PRINT_ARTIFACT_TTL", "168h"), TerminalTTL: envOrDefault("STUFF_STASH_PRINT_TERMINAL_TTL", "720h"), Lease: envOrDefault("STUFF_STASH_PRINT_CLAIM_LEASE", "60s"), ReadinessMaxAge: envOrDefault("STUFF_STASH_PRINT_READINESS_MAX_AGE", "90s"), CleanupInterval: envOrDefault("STUFF_STASH_PRINT_CLEANUP_INTERVAL", "1m")}
}
func (c PrintJobConfig) Settings() (printingapp.JobConfig, error) {
	out := printingapp.JobConfig{MaxCopies: 20, MaxArtifactBytes: 1048576, ArtifactTTL: 7 * 24 * time.Hour, TerminalTTL: 30 * 24 * time.Hour, Lease: time.Minute, ReadinessMaxAge: 90 * time.Second, CleanupInterval: time.Minute}
	for _, v := range []struct {
		raw    string
		target *int
		max    int
	}{{c.MaxCopies, &out.MaxCopies, 1000}, {c.MaxArtifactBytes, &out.MaxArtifactBytes, 16 * 1024 * 1024}} {
		if v.raw != "" {
			n, e := strconv.Atoi(v.raw)
			if e != nil || n < 1 || n > v.max {
				return out, errors.New("invalid print job size limit")
			}
			*v.target = n
		}
	}
	for _, v := range []struct {
		raw    string
		target *time.Duration
		max    time.Duration
	}{{c.ArtifactTTL, &out.ArtifactTTL, 30 * 24 * time.Hour}, {c.TerminalTTL, &out.TerminalTTL, 365 * 24 * time.Hour}, {c.Lease, &out.Lease, 10 * time.Minute}, {c.ReadinessMaxAge, &out.ReadinessMaxAge, time.Hour}, {c.CleanupInterval, &out.CleanupInterval, time.Hour}} {
		if v.raw != "" {
			d, e := time.ParseDuration(v.raw)
			if e != nil || d <= 0 || d > v.max {
				return out, errors.New("invalid print job duration")
			}
			*v.target = d
		}
	}
	if out.TerminalTTL < out.ArtifactTTL {
		return out, errors.New("print history retention must cover artifact lifetime")
	}
	return out, nil
}
