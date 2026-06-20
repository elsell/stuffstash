package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envHTTPAddr                 = "STUFF_STASH_HTTP_ADDR"
	envCORSAllowedOrigins       = "STUFF_STASH_CORS_ALLOWED_ORIGINS"
	envAuthMode                 = "STUFF_STASH_AUTH_MODE"
	envAuthzMode                = "STUFF_STASH_AUTHZ_MODE"
	envOIDCIssuer               = "STUFF_STASH_OIDC_ISSUER"
	envOIDCClientID             = "STUFF_STASH_OIDC_CLIENT_ID"
	envOIDCClientIDs            = "STUFF_STASH_OIDC_CLIENT_IDS"
	envRepositoryMode           = "STUFF_STASH_REPOSITORY_MODE"
	envDatabaseDSN              = "STUFF_STASH_DATABASE_DSN"
	envSpiceDBEndpoint          = "STUFF_STASH_SPICEDB_ENDPOINT"
	envSpiceDBPresharedKey      = "STUFF_STASH_SPICEDB_PRESHARED_KEY"
	envSpiceDBTLSEnabled        = "STUFF_STASH_SPICEDB_TLS_ENABLED"
	envSpiceDBBootstrapSchema   = "STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA"
	envSpiceDBSchemaPath        = "STUFF_STASH_SPICEDB_SCHEMA_PATH"
	envAuthorizationOutboxLimit = "STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_LIMIT"
	envAuthorizationOutboxEvery = "STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_INTERVAL"
	envAuthorizationOutboxLease = "STUFF_STASH_AUTHORIZATION_OUTBOX_CLAIM_LEASE"
	envInvitationTTL            = "STUFF_STASH_INVITATION_TTL"
	envDefaultPageLimit         = "STUFF_STASH_DEFAULT_PAGE_LIMIT"
	envMaxPageLimit             = "STUFF_STASH_MAX_PAGE_LIMIT"
	envBlobStorageMode          = "STUFF_STASH_BLOB_STORAGE_MODE"
	envBlobStoragePath          = "STUFF_STASH_BLOB_STORAGE_PATH"
	envS3Endpoint               = "STUFF_STASH_S3_ENDPOINT"
	envS3AccessKey              = "STUFF_STASH_S3_ACCESS_KEY"
	envS3SecretKey              = "STUFF_STASH_S3_SECRET_KEY"
	envS3Bucket                 = "STUFF_STASH_S3_BUCKET"
	envS3Region                 = "STUFF_STASH_S3_REGION"
	envS3Secure                 = "STUFF_STASH_S3_SECURE"
	envMaxAttachmentBytes       = "STUFF_STASH_MAX_ATTACHMENT_BYTES"
	defaultHTTPAddr             = ":8080"
	defaultAuthMode             = "local-dev"
	defaultAuthzMode            = "memory"
	defaultRepositoryMode       = "memory"
	defaultSpiceDBSchemaPath    = "deploy/spicedb/schema.zed"
	defaultAuthorizationLimit   = 25
	defaultAuthorizationEvery   = 10 * time.Second
	defaultAuthorizationLease   = 30 * time.Second
	defaultInvitationTTL        = 7 * 24 * time.Hour
	defaultDefaultPageLimit     = 50
	defaultMaxPageLimit         = 100
	defaultBlobStorageMode      = "filesystem"
	defaultBlobStoragePath      = ".stuffstash/blobs"
	defaultS3Region             = "garage"
	defaultS3Secure             = true
	defaultMaxAttachmentBytes   = 5 * 1024 * 1024
	defaultSpiceDBTLSEnabled    = true
	defaultSpiceDBBootstrapMode = false
)

type Config struct {
	HTTPAddr                         string
	CORSAllowedOrigins               []string
	AuthMode                         string
	AuthzMode                        string
	OIDCIssuer                       string
	OIDCClientID                     string
	OIDCClientIDs                    []string
	RepositoryMode                   string
	DatabaseDSN                      string
	SpiceDBEndpoint                  string
	SpiceDBPresharedKey              string
	SpiceDBTLSEnabled                bool
	SpiceDBBootstrapSchema           bool
	SpiceDBSchemaPath                string
	AuthorizationOutboxDrainLimit    int
	AuthorizationOutboxDrainInterval time.Duration
	AuthorizationOutboxClaimLease    time.Duration
	InvitationTTL                    time.Duration
	DefaultPageLimit                 int
	MaxPageLimit                     int
	BlobStorageMode                  string
	BlobStoragePath                  string
	S3Endpoint                       string
	S3AccessKey                      string
	S3SecretKey                      string
	S3Bucket                         string
	S3Region                         string
	S3Secure                         bool
	MaxAttachmentBytes               int
}

func Load() Config {
	return Config{
		HTTPAddr:                         envOrDefault(envHTTPAddr, defaultHTTPAddr),
		CORSAllowedOrigins:               stringListEnv(envCORSAllowedOrigins),
		AuthMode:                         envOrDefault(envAuthMode, defaultAuthMode),
		AuthzMode:                        envOrDefault(envAuthzMode, defaultAuthzMode),
		OIDCIssuer:                       os.Getenv(envOIDCIssuer),
		OIDCClientID:                     os.Getenv(envOIDCClientID),
		OIDCClientIDs:                    oidcClientIDs(),
		RepositoryMode:                   envOrDefault(envRepositoryMode, defaultRepositoryMode),
		DatabaseDSN:                      os.Getenv(envDatabaseDSN),
		SpiceDBEndpoint:                  os.Getenv(envSpiceDBEndpoint),
		SpiceDBPresharedKey:              os.Getenv(envSpiceDBPresharedKey),
		SpiceDBTLSEnabled:                boolEnvOrDefault(envSpiceDBTLSEnabled, defaultSpiceDBTLSEnabled),
		SpiceDBBootstrapSchema:           boolEnvOrDefault(envSpiceDBBootstrapSchema, defaultSpiceDBBootstrapMode),
		SpiceDBSchemaPath:                envOrDefault(envSpiceDBSchemaPath, defaultSpiceDBSchemaPath),
		AuthorizationOutboxDrainLimit:    intEnvOrDefault(envAuthorizationOutboxLimit, defaultAuthorizationLimit),
		AuthorizationOutboxDrainInterval: durationEnvOrDefault(envAuthorizationOutboxEvery, defaultAuthorizationEvery),
		AuthorizationOutboxClaimLease:    durationEnvOrDefault(envAuthorizationOutboxLease, defaultAuthorizationLease),
		InvitationTTL:                    durationEnvOrDefault(envInvitationTTL, defaultInvitationTTL),
		DefaultPageLimit:                 intEnvOrDefault(envDefaultPageLimit, defaultDefaultPageLimit),
		MaxPageLimit:                     intEnvOrDefault(envMaxPageLimit, defaultMaxPageLimit),
		BlobStorageMode:                  envOrDefault(envBlobStorageMode, defaultBlobStorageMode),
		BlobStoragePath:                  envOrDefault(envBlobStoragePath, defaultBlobStoragePath),
		S3Endpoint:                       os.Getenv(envS3Endpoint),
		S3AccessKey:                      os.Getenv(envS3AccessKey),
		S3SecretKey:                      os.Getenv(envS3SecretKey),
		S3Bucket:                         os.Getenv(envS3Bucket),
		S3Region:                         envOrDefault(envS3Region, defaultS3Region),
		S3Secure:                         boolEnvOrDefault(envS3Secure, defaultS3Secure),
		MaxAttachmentBytes:               intEnvOrDefault(envMaxAttachmentBytes, defaultMaxAttachmentBytes),
	}
}

func oidcClientIDs() []string {
	clientIDs := stringListEnv(envOIDCClientIDs)
	singleClientID := strings.TrimSpace(os.Getenv(envOIDCClientID))
	if singleClientID == "" {
		return clientIDs
	}
	for _, clientID := range clientIDs {
		if clientID == singleClientID {
			return clientIDs
		}
	}
	return append([]string{singleClientID}, clientIDs...)
}

func stringListEnv(name string) []string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		items = append(items, item)
	}
	return items
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
