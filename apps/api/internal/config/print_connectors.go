package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type PrintConnectorConfig struct {
	PublicWebBaseURL                                                                                          string
	PairingLifetime, CredentialLifetime, ActivationLifetime, AuthorizationTimeout, PollInterval, ReportMaxAge time.Duration
	BatchSize                                                                                                 int
}

func LoadPrintConnectors() (PrintConnectorConfig, error) {
	c := PrintConnectorConfig{PublicWebBaseURL: strings.TrimRight(os.Getenv("STUFF_STASH_PUBLIC_WEB_BASE_URL"), "/"), PairingLifetime: 10 * time.Minute, CredentialLifetime: 30 * 24 * time.Hour, ActivationLifetime: 5 * time.Minute, AuthorizationTimeout: 5 * time.Second, PollInterval: 5 * time.Second, ReportMaxAge: time.Minute, BatchSize: 100}
	for _, field := range []struct {
		name   string
		target *time.Duration
	}{{"PAIRING_LIFETIME", &c.PairingLifetime}, {"CREDENTIAL_LIFETIME", &c.CredentialLifetime}, {"ACTIVATION_LIFETIME", &c.ActivationLifetime}, {"AUTHORIZATION_TIMEOUT", &c.AuthorizationTimeout}, {"POLL_INTERVAL", &c.PollInterval}, {"REPORT_MAX_AGE", &c.ReportMaxAge}} {
		if raw, ok := os.LookupEnv("STUFF_STASH_PRINT_CONNECTOR_" + field.name); ok {
			value, err := time.ParseDuration(raw)
			if err != nil {
				return c, fmt.Errorf("invalid print connector %s", field.name)
			}
			*field.target = value
		}
	}
	if raw, ok := os.LookupEnv("STUFF_STASH_PRINT_CONNECTOR_BATCH_SIZE"); ok {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return c, fmt.Errorf("invalid print connector batch size")
		}
		c.BatchSize = value
	}
	if c.PairingLifetime < time.Minute || c.PairingLifetime > time.Hour || c.CredentialLifetime < time.Hour || c.ActivationLifetime < time.Second || c.ActivationLifetime > c.CredentialLifetime || c.AuthorizationTimeout < time.Millisecond || c.AuthorizationTimeout > time.Minute || c.PollInterval < 100*time.Millisecond || c.BatchSize < 1 || c.BatchSize > 1000 || c.ReportMaxAge < time.Second {
		return c, fmt.Errorf("invalid print connector policy bounds")
	}
	if c.PublicWebBaseURL != "" {
		u, err := url.Parse(c.PublicWebBaseURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return c, fmt.Errorf("invalid print connector public web base URL")
		}
	}
	return c, nil
}
