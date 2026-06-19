package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envHTTPAddr                 = "STUFF_STASH_HTTP_ADDR"
	envAuthMode                 = "STUFF_STASH_AUTH_MODE"
	envAuthzMode                = "STUFF_STASH_AUTHZ_MODE"
	envOIDCIssuer               = "STUFF_STASH_OIDC_ISSUER"
	envOIDCClientID             = "STUFF_STASH_OIDC_CLIENT_ID"
	envRepositoryMode           = "STUFF_STASH_REPOSITORY_MODE"
	envDatabaseDSN              = "STUFF_STASH_DATABASE_DSN"
	envSpiceDBEndpoint          = "STUFF_STASH_SPICEDB_ENDPOINT"
	envSpiceDBPresharedKey      = "STUFF_STASH_SPICEDB_PRESHARED_KEY"
	envSpiceDBTLSEnabled        = "STUFF_STASH_SPICEDB_TLS_ENABLED"
	envSpiceDBBootstrapSchema   = "STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA"
	envSpiceDBSchemaPath        = "STUFF_STASH_SPICEDB_SCHEMA_PATH"
	envAuthorizationOutboxLimit = "STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_LIMIT"
	envAuthorizationOutboxEvery = "STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_INTERVAL"
	defaultHTTPAddr             = ":8080"
	defaultAuthMode             = "local-dev"
	defaultAuthzMode            = "memory"
	defaultRepositoryMode       = "memory"
	defaultSpiceDBSchemaPath    = "deploy/spicedb/schema.zed"
	defaultAuthorizationLimit   = 25
	defaultAuthorizationEvery   = 10 * time.Second
	defaultSpiceDBTLSEnabled    = true
	defaultSpiceDBBootstrapMode = false
)

type Config struct {
	HTTPAddr                         string
	AuthMode                         string
	AuthzMode                        string
	OIDCIssuer                       string
	OIDCClientID                     string
	RepositoryMode                   string
	DatabaseDSN                      string
	SpiceDBEndpoint                  string
	SpiceDBPresharedKey              string
	SpiceDBTLSEnabled                bool
	SpiceDBBootstrapSchema           bool
	SpiceDBSchemaPath                string
	AuthorizationOutboxDrainLimit    int
	AuthorizationOutboxDrainInterval time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:                         envOrDefault(envHTTPAddr, defaultHTTPAddr),
		AuthMode:                         envOrDefault(envAuthMode, defaultAuthMode),
		AuthzMode:                        envOrDefault(envAuthzMode, defaultAuthzMode),
		OIDCIssuer:                       os.Getenv(envOIDCIssuer),
		OIDCClientID:                     os.Getenv(envOIDCClientID),
		RepositoryMode:                   envOrDefault(envRepositoryMode, defaultRepositoryMode),
		DatabaseDSN:                      os.Getenv(envDatabaseDSN),
		SpiceDBEndpoint:                  os.Getenv(envSpiceDBEndpoint),
		SpiceDBPresharedKey:              os.Getenv(envSpiceDBPresharedKey),
		SpiceDBTLSEnabled:                boolEnvOrDefault(envSpiceDBTLSEnabled, defaultSpiceDBTLSEnabled),
		SpiceDBBootstrapSchema:           boolEnvOrDefault(envSpiceDBBootstrapSchema, defaultSpiceDBBootstrapMode),
		SpiceDBSchemaPath:                envOrDefault(envSpiceDBSchemaPath, defaultSpiceDBSchemaPath),
		AuthorizationOutboxDrainLimit:    intEnvOrDefault(envAuthorizationOutboxLimit, defaultAuthorizationLimit),
		AuthorizationOutboxDrainInterval: durationEnvOrDefault(envAuthorizationOutboxEvery, defaultAuthorizationEvery),
	}
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func intEnvOrDefault(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func durationEnvOrDefault(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func boolEnvOrDefault(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	switch value {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
