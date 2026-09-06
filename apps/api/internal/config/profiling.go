package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type ProfilingConfig struct {
	Enabled                                  bool
	Endpoint, Username, Password             string
	ServiceName, ServiceVersion, Environment string
	UploadInterval, RequestTimeout           time.Duration
	MutexFraction, BlockRate                 int
}

func LoadProfiling() (ProfilingConfig, error) {
	cfg := ProfilingConfig{ServiceName: "stuffstash-api", UploadInterval: 15 * time.Second, RequestTimeout: 5 * time.Second, MutexFraction: 5, BlockRate: 1000000}
	var err error
	if raw := os.Getenv("STUFF_STASH_PROFILING_ENABLED"); raw != "" {
		cfg.Enabled, err = strconv.ParseBool(raw)
		if err != nil {
			return cfg, errors.New("invalid profiling enabled setting")
		}
	}
	if !cfg.Enabled {
		return cfg, nil
	}
	cfg.Endpoint = os.Getenv("STUFF_STASH_PROFILING_ENDPOINT")
	cfg.Username = os.Getenv("STUFF_STASH_PROFILING_USERNAME")
	cfg.Password = os.Getenv("STUFF_STASH_PROFILING_PASSWORD")
	if value := os.Getenv("OTEL_SERVICE_NAME"); value != "" {
		cfg.ServiceName = value
	}
	cfg.ServiceVersion = os.Getenv("OTEL_SERVICE_VERSION")
	cfg.Environment = os.Getenv("STUFF_STASH_DEPLOYMENT_ENVIRONMENT")
	for name, target := range map[string]*time.Duration{"STUFF_STASH_PROFILING_UPLOAD_INTERVAL": &cfg.UploadInterval, "STUFF_STASH_PROFILING_REQUEST_TIMEOUT": &cfg.RequestTimeout} {
		if raw := os.Getenv(name); raw != "" {
			*target, err = time.ParseDuration(raw)
			if err != nil {
				return cfg, errors.New("invalid profiling duration")
			}
		}
	}
	for name, target := range map[string]*int{"STUFF_STASH_PROFILING_MUTEX_FRACTION": &cfg.MutexFraction, "STUFF_STASH_PROFILING_BLOCK_RATE": &cfg.BlockRate} {
		if raw := os.Getenv(name); raw != "" {
			*target, err = strconv.Atoi(raw)
			if err != nil {
				return cfg, errors.New("invalid profiling sample setting")
			}
		}
	}
	return cfg, cfg.Validate()
}

func (c ProfilingConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if _, present := os.LookupEnv("PYROSCOPE_ADHOC_SERVER_ADDRESS"); present {
		return errors.New("profiling endpoint override is unsupported")
	}
	if !validTelemetryEndpoint(c.Endpoint) {
		return errors.New("invalid profiling endpoint")
	}
	if c.UploadInterval <= 0 || c.RequestTimeout <= 0 {
		return errors.New("invalid profiling duration")
	}
	if c.MutexFraction < 0 || c.BlockRate < 0 {
		return errors.New("invalid profiling sample setting")
	}
	if c.ServiceName == "" || len(c.ServiceName) > 128 || len(c.ServiceVersion) > 128 || len(c.Environment) > 128 {
		return errors.New("invalid profiling resource identity")
	}
	for _, value := range []string{c.ServiceName, c.ServiceVersion, c.Environment} {
		for _, char := range value {
			if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("._-", char)) {
				return errors.New("invalid profiling resource identity")
			}
		}
	}
	return nil
}
