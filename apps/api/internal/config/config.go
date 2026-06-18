package config

import (
	"os"
	"strings"
)

const (
	envHTTPAddr                 = "STUFF_STASH_HTTP_ADDR"
	envAuthMode                 = "STUFF_STASH_AUTH_MODE"
	envAuthzMode                = "STUFF_STASH_AUTHZ_MODE"
	envOIDCIssuer               = "STUFF_STASH_OIDC_ISSUER"
	envOIDCClientID             = "STUFF_STASH_OIDC_CLIENT_ID"
	envSpiceDBEndpoint          = "STUFF_STASH_SPICEDB_ENDPOINT"
	envSpiceDBPresharedKey      = "STUFF_STASH_SPICEDB_PRESHARED_KEY"
	envSpiceDBTLSEnabled        = "STUFF_STASH_SPICEDB_TLS_ENABLED"
	envSpiceDBBootstrapSchema   = "STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA"
	envSpiceDBSchemaPath        = "STUFF_STASH_SPICEDB_SCHEMA_PATH"
	defaultHTTPAddr             = ":8080"
	defaultAuthMode             = "local-dev"
	defaultAuthzMode            = "memory"
	defaultSpiceDBSchemaPath    = "deploy/spicedb/schema.zed"
	defaultSpiceDBTLSEnabled    = true
	defaultSpiceDBBootstrapMode = false
)

type Config struct {
	HTTPAddr               string
	AuthMode               string
	AuthzMode              string
	OIDCIssuer             string
	OIDCClientID           string
	SpiceDBEndpoint        string
	SpiceDBPresharedKey    string
	SpiceDBTLSEnabled      bool
	SpiceDBBootstrapSchema bool
	SpiceDBSchemaPath      string
}

func Load() Config {
	return Config{
		HTTPAddr:               envOrDefault(envHTTPAddr, defaultHTTPAddr),
		AuthMode:               envOrDefault(envAuthMode, defaultAuthMode),
		AuthzMode:              envOrDefault(envAuthzMode, defaultAuthzMode),
		OIDCIssuer:             os.Getenv(envOIDCIssuer),
		OIDCClientID:           os.Getenv(envOIDCClientID),
		SpiceDBEndpoint:        os.Getenv(envSpiceDBEndpoint),
		SpiceDBPresharedKey:    os.Getenv(envSpiceDBPresharedKey),
		SpiceDBTLSEnabled:      boolEnvOrDefault(envSpiceDBTLSEnabled, defaultSpiceDBTLSEnabled),
		SpiceDBBootstrapSchema: boolEnvOrDefault(envSpiceDBBootstrapSchema, defaultSpiceDBBootstrapMode),
		SpiceDBSchemaPath:      envOrDefault(envSpiceDBSchemaPath, defaultSpiceDBSchemaPath),
	}
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
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
