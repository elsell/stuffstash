package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultHTTPRateLimitRequests          = 1200
	DefaultHTTPRateLimitBurst             = 600
	envHTTPAddr                           = "STUFF_STASH_HTTP_ADDR"
	envHTTPReadHeaderTimeout              = "STUFF_STASH_HTTP_READ_HEADER_TIMEOUT"
	envHTTPReadTimeout                    = "STUFF_STASH_HTTP_READ_TIMEOUT"
	envHTTPWriteTimeout                   = "STUFF_STASH_HTTP_WRITE_TIMEOUT"
	envHTTPIdleTimeout                    = "STUFF_STASH_HTTP_IDLE_TIMEOUT"
	envHTTPMaxJSONBodyBytes               = "STUFF_STASH_HTTP_MAX_JSON_BODY_BYTES"
	envHTTPRateLimitEnabled               = "STUFF_STASH_HTTP_RATE_LIMIT_ENABLED"
	envHTTPRateLimitRequests              = "STUFF_STASH_HTTP_RATE_LIMIT_REQUESTS"
	envHTTPRateLimitWindow                = "STUFF_STASH_HTTP_RATE_LIMIT_WINDOW"
	envHTTPRateLimitBurst                 = "STUFF_STASH_HTTP_RATE_LIMIT_BURST"
	envCORSAllowedOrigins                 = "STUFF_STASH_CORS_ALLOWED_ORIGINS"
	envAuthMode                           = "STUFF_STASH_AUTH_MODE"
	envAuthzMode                          = "STUFF_STASH_AUTHZ_MODE"
	envOIDCIssuer                         = "STUFF_STASH_OIDC_ISSUER"
	envOIDCClientID                       = "STUFF_STASH_OIDC_CLIENT_ID"
	envOIDCClientIDs                      = "STUFF_STASH_OIDC_CLIENT_IDS"
	envOIDCMobileClientID                 = "STUFF_STASH_OIDC_MOBILE_CLIENT_ID"
	envOIDCMobileRedirectURI              = "STUFF_STASH_OIDC_MOBILE_REDIRECT_URI"
	envOIDCMobileScopes                   = "STUFF_STASH_OIDC_MOBILE_SCOPES"
	envRepositoryMode                     = "STUFF_STASH_REPOSITORY_MODE"
	envDatabaseDSN                        = "STUFF_STASH_DATABASE_DSN"
	envSpiceDBEndpoint                    = "STUFF_STASH_SPICEDB_ENDPOINT"
	envSpiceDBPresharedKey                = "STUFF_STASH_SPICEDB_PRESHARED_KEY"
	envSpiceDBTLSEnabled                  = "STUFF_STASH_SPICEDB_TLS_ENABLED"
	envSpiceDBCAPath                      = "STUFF_STASH_SPICEDB_CA_PATH"
	envSpiceDBBootstrapSchema             = "STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA"
	envSpiceDBSchemaPath                  = "STUFF_STASH_SPICEDB_SCHEMA_PATH"
	envAuthorizationOutboxLimit           = "STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_LIMIT"
	envAuthorizationOutboxEvery           = "STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_INTERVAL"
	envAuthorizationOutboxLease           = "STUFF_STASH_AUTHORIZATION_OUTBOX_CLAIM_LEASE"
	envBlobDeletionOutboxLimit            = "STUFF_STASH_BLOB_DELETION_OUTBOX_DRAIN_LIMIT"
	envBlobDeletionOutboxEvery            = "STUFF_STASH_BLOB_DELETION_OUTBOX_DRAIN_INTERVAL"
	envBlobDeletionOutboxLease            = "STUFF_STASH_BLOB_DELETION_OUTBOX_CLAIM_LEASE"
	envBlobDeletionOutboxMaxAttempts      = "STUFF_STASH_BLOB_DELETION_OUTBOX_MAX_ATTEMPTS"
	envInvitationTTL                      = "STUFF_STASH_INVITATION_TTL"
	envDefaultPageLimit                   = "STUFF_STASH_DEFAULT_PAGE_LIMIT"
	envMaxPageLimit                       = "STUFF_STASH_MAX_PAGE_LIMIT"
	envBlobStorageMode                    = "STUFF_STASH_BLOB_STORAGE_MODE"
	envBlobStoragePath                    = "STUFF_STASH_BLOB_STORAGE_PATH"
	envS3Endpoint                         = "STUFF_STASH_S3_ENDPOINT"
	envS3PublicEndpoint                   = "STUFF_STASH_S3_PUBLIC_ENDPOINT"
	envS3AccessKey                        = "STUFF_STASH_S3_ACCESS_KEY"
	envS3SecretKey                        = "STUFF_STASH_S3_SECRET_KEY"
	envS3Bucket                           = "STUFF_STASH_S3_BUCKET"
	envS3Region                           = "STUFF_STASH_S3_REGION"
	envS3Secure                           = "STUFF_STASH_S3_SECURE"
	envMaxAttachmentBytes                 = "STUFF_STASH_MAX_ATTACHMENT_BYTES"
	envPrimaryThumbnailWarmLimit          = "STUFF_STASH_PRIMARY_THUMBNAIL_WARM_LIMIT"
	envPrimaryThumbnailWarmConcurrent     = "STUFF_STASH_PRIMARY_THUMBNAIL_WARM_CONCURRENCY"
	envPrimaryThumbnailWarmTimeout        = "STUFF_STASH_PRIMARY_THUMBNAIL_WARM_TIMEOUT"
	envVoiceDevFakeEnabled                = "STUFF_STASH_VOICE_DEV_FAKE_ENABLED"
	envVoiceGoogleEnabled                 = "STUFF_STASH_VOICE_GOOGLE_ENABLED"
	envRealtimeVoiceIdleTimeout           = "STUFF_STASH_REALTIME_VOICE_IDLE_TIMEOUT"
	envVoiceProviderHTTPTimeout           = "STUFF_STASH_VOICE_PROVIDER_HTTP_TIMEOUT"
	envGoogleCloudProject                 = "STUFF_STASH_GOOGLE_CLOUD_PROJECT"
	envGoogleCloudLocation                = "STUFF_STASH_GOOGLE_CLOUD_LOCATION"
	envGoogleGeminiModel                  = "STUFF_STASH_GOOGLE_GEMINI_MODEL"
	envGoogleTTSLanguageCode              = "STUFF_STASH_GOOGLE_TTS_LANGUAGE_CODE"
	envGoogleTTSVoiceName                 = "STUFF_STASH_GOOGLE_TTS_VOICE_NAME"
	envGoogleCredentialMode               = "STUFF_STASH_GOOGLE_CREDENTIAL_MODE"
	envGoogleAccessToken                  = "STUFF_STASH_GOOGLE_ACCESS_TOKEN"
	envProviderCredentialKeyID            = "STUFF_STASH_PROVIDER_CREDENTIAL_KEY_ID"
	envProviderCredentialKey              = "STUFF_STASH_PROVIDER_CREDENTIAL_KEY"
	envImportJobTimeoutSeconds            = "STUFF_STASH_IMPORT_JOB_TIMEOUT_SECONDS"
	envImportCredentialVacuumSeconds      = "STUFF_STASH_IMPORT_CREDENTIAL_VACUUM_INTERVAL_SECONDS"
	defaultHTTPAddr                       = ":8080"
	defaultHTTPReadHeader                 = 5 * time.Second
	defaultHTTPRead                       = 15 * time.Second
	defaultHTTPWrite                      = 30 * time.Second
	defaultHTTPIdle                       = 60 * time.Second
	defaultHTTPMaxJSONBodyBytes           = 1024 * 1024
	defaultHTTPRateLimitEnabled           = true
	defaultHTTPRateLimitRequests          = DefaultHTTPRateLimitRequests
	defaultHTTPRateLimitWindow            = time.Minute
	defaultHTTPRateLimitBurst             = DefaultHTTPRateLimitBurst
	defaultAuthMode                       = "local-dev"
	defaultAuthzMode                      = "memory"
	defaultOIDCMobileRedirectURI          = "stuffstash://auth/callback"
	defaultRepositoryMode                 = "memory"
	defaultSpiceDBSchemaPath              = "deploy/spicedb/schema.zed"
	defaultAuthorizationLimit             = 25
	defaultAuthorizationEvery             = 10 * time.Second
	defaultAuthorizationLease             = 30 * time.Second
	defaultBlobDeletionLimit              = 25
	defaultBlobDeletionEvery              = 10 * time.Second
	defaultBlobDeletionLease              = 30 * time.Second
	defaultBlobDeletionMaxAttempts        = 5
	defaultInvitationTTL                  = 7 * 24 * time.Hour
	defaultDefaultPageLimit               = 50
	defaultMaxPageLimit                   = 100
	defaultBlobStorageMode                = "filesystem"
	defaultBlobStoragePath                = ".stuffstash/blobs"
	defaultS3Region                       = "garage"
	defaultS3Secure                       = true
	defaultMaxAttachmentBytes             = 25 * 1024 * 1024
	defaultPrimaryThumbnailWarmLimit      = 12
	defaultPrimaryThumbnailWarmConcurrent = 4
	defaultPrimaryThumbnailWarmTimeout    = 10 * time.Second
	defaultSpiceDBTLSEnabled              = true
	defaultSpiceDBBootstrapMode           = false
	defaultVoiceDevFakeEnabled            = false
	defaultVoiceGoogleEnabled             = false
	defaultRealtimeVoiceIdleTimeout       = 15 * time.Second
	defaultVoiceProviderHTTPTimeout       = 60 * time.Second
	defaultGoogleCloudLocation            = "us-central1"
	defaultGoogleGeminiModel              = "gemini-2.5-flash-lite"
	defaultGoogleTTSLanguageCode          = "en-US"
	defaultGoogleTTSVoiceName             = "en-US-Standard-C"
	defaultImportJobTimeout               = 15 * time.Minute
	defaultImportCredentialVacuumInterval = time.Minute
)

const (
	GoogleCredentialModeADC         = "adc"
	GoogleCredentialModeAccessToken = "access_token"
)

const defaultGoogleCredentialMode = GoogleCredentialModeADC

type Config struct {
	HTTPAddr                         string
	HTTPReadHeaderTimeout            time.Duration
	HTTPReadTimeout                  time.Duration
	HTTPWriteTimeout                 time.Duration
	HTTPIdleTimeout                  time.Duration
	HTTPMaxJSONBodyBytes             int64
	HTTPRateLimitEnabled             bool
	HTTPRateLimitRequests            int
	HTTPRateLimitWindow              time.Duration
	HTTPRateLimitBurst               int
	CORSAllowedOrigins               []string
	AuthMode                         string
	AuthzMode                        string
	OIDCIssuer                       string
	OIDCClientID                     string
	OIDCClientIDs                    []string
	OIDCMobileClientID               string
	OIDCMobileRedirectURI            string
	OIDCMobileScopes                 []string
	RepositoryMode                   string
	DatabaseDSN                      string
	SpiceDBEndpoint                  string
	SpiceDBPresharedKey              string
	SpiceDBTLSEnabled                bool
	SpiceDBCAPath                    string
	SpiceDBBootstrapSchema           bool
	SpiceDBSchemaPath                string
	AuthorizationOutboxDrainLimit    int
	AuthorizationOutboxDrainInterval time.Duration
	AuthorizationOutboxClaimLease    time.Duration
	BlobDeletionOutboxDrainLimit     int
	BlobDeletionOutboxDrainInterval  time.Duration
	BlobDeletionOutboxClaimLease     time.Duration
	BlobDeletionOutboxMaxAttempts    int
	InvitationTTL                    time.Duration
	DefaultPageLimit                 int
	MaxPageLimit                     int
	BlobStorageMode                  string
	BlobStoragePath                  string
	S3Endpoint                       string
	S3PublicEndpoint                 string
	S3AccessKey                      string
	S3SecretKey                      string
	S3Bucket                         string
	S3Region                         string
	S3Secure                         bool
	MaxAttachmentBytes               int
	PrimaryThumbnailWarmLimit        int
	PrimaryThumbnailWarmConcurrency  int
	PrimaryThumbnailWarmTimeout      time.Duration
	VoiceDevFakeEnabled              bool
	VoiceGoogleEnabled               bool
	RealtimeVoiceIdleTimeout         time.Duration
	VoiceProviderHTTPTimeout         time.Duration
	GoogleCloudProject               string
	GoogleCloudLocation              string
	GoogleGeminiModel                string
	GoogleTTSLanguageCode            string
	GoogleTTSVoiceName               string
	GoogleCredentialMode             string
	GoogleAccessToken                string
	ProviderCredentialKeyID          string
	ProviderCredentialKey            string
	ImportJobTimeout                 time.Duration
	ImportCredentialVacuumInterval   time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:                         envOrDefault(envHTTPAddr, defaultHTTPAddr),
		HTTPReadHeaderTimeout:            durationEnvOrDefault(envHTTPReadHeaderTimeout, defaultHTTPReadHeader),
		HTTPReadTimeout:                  durationEnvOrDefault(envHTTPReadTimeout, defaultHTTPRead),
		HTTPWriteTimeout:                 durationEnvOrDefault(envHTTPWriteTimeout, defaultHTTPWrite),
		HTTPIdleTimeout:                  durationEnvOrDefault(envHTTPIdleTimeout, defaultHTTPIdle),
		HTTPMaxJSONBodyBytes:             int64EnvOrDefault(envHTTPMaxJSONBodyBytes, defaultHTTPMaxJSONBodyBytes),
		HTTPRateLimitEnabled:             boolEnvOrDefault(envHTTPRateLimitEnabled, defaultHTTPRateLimitEnabled),
		HTTPRateLimitRequests:            intEnvOrDefault(envHTTPRateLimitRequests, defaultHTTPRateLimitRequests),
		HTTPRateLimitWindow:              durationEnvOrDefault(envHTTPRateLimitWindow, defaultHTTPRateLimitWindow),
		HTTPRateLimitBurst:               intEnvOrDefault(envHTTPRateLimitBurst, defaultHTTPRateLimitBurst),
		CORSAllowedOrigins:               stringListEnv(envCORSAllowedOrigins),
		AuthMode:                         envOrDefault(envAuthMode, defaultAuthMode),
		AuthzMode:                        envOrDefault(envAuthzMode, defaultAuthzMode),
		OIDCIssuer:                       os.Getenv(envOIDCIssuer),
		OIDCClientID:                     os.Getenv(envOIDCClientID),
		OIDCClientIDs:                    oidcClientIDs(),
		OIDCMobileClientID:               strings.TrimSpace(os.Getenv(envOIDCMobileClientID)),
		OIDCMobileRedirectURI:            envOrDefault(envOIDCMobileRedirectURI, defaultOIDCMobileRedirectURI),
		OIDCMobileScopes:                 oidcMobileScopes(),
		RepositoryMode:                   envOrDefault(envRepositoryMode, defaultRepositoryMode),
		DatabaseDSN:                      os.Getenv(envDatabaseDSN),
		SpiceDBEndpoint:                  os.Getenv(envSpiceDBEndpoint),
		SpiceDBPresharedKey:              os.Getenv(envSpiceDBPresharedKey),
		SpiceDBTLSEnabled:                boolEnvOrDefault(envSpiceDBTLSEnabled, defaultSpiceDBTLSEnabled),
		SpiceDBCAPath:                    os.Getenv(envSpiceDBCAPath),
		SpiceDBBootstrapSchema:           boolEnvOrDefault(envSpiceDBBootstrapSchema, defaultSpiceDBBootstrapMode),
		SpiceDBSchemaPath:                envOrDefault(envSpiceDBSchemaPath, defaultSpiceDBSchemaPath),
		AuthorizationOutboxDrainLimit:    intEnvOrDefault(envAuthorizationOutboxLimit, defaultAuthorizationLimit),
		AuthorizationOutboxDrainInterval: durationEnvOrDefault(envAuthorizationOutboxEvery, defaultAuthorizationEvery),
		AuthorizationOutboxClaimLease:    durationEnvOrDefault(envAuthorizationOutboxLease, defaultAuthorizationLease),
		BlobDeletionOutboxDrainLimit:     intEnvOrDefault(envBlobDeletionOutboxLimit, defaultBlobDeletionLimit),
		BlobDeletionOutboxDrainInterval:  durationEnvOrDefault(envBlobDeletionOutboxEvery, defaultBlobDeletionEvery),
		BlobDeletionOutboxClaimLease:     durationEnvOrDefault(envBlobDeletionOutboxLease, defaultBlobDeletionLease),
		BlobDeletionOutboxMaxAttempts:    intEnvOrDefault(envBlobDeletionOutboxMaxAttempts, defaultBlobDeletionMaxAttempts),
		InvitationTTL:                    durationEnvOrDefault(envInvitationTTL, defaultInvitationTTL),
		DefaultPageLimit:                 intEnvOrDefault(envDefaultPageLimit, defaultDefaultPageLimit),
		MaxPageLimit:                     intEnvOrDefault(envMaxPageLimit, defaultMaxPageLimit),
		BlobStorageMode:                  envOrDefault(envBlobStorageMode, defaultBlobStorageMode),
		BlobStoragePath:                  envOrDefault(envBlobStoragePath, defaultBlobStoragePath),
		S3Endpoint:                       os.Getenv(envS3Endpoint),
		S3PublicEndpoint:                 os.Getenv(envS3PublicEndpoint),
		S3AccessKey:                      os.Getenv(envS3AccessKey),
		S3SecretKey:                      os.Getenv(envS3SecretKey),
		S3Bucket:                         os.Getenv(envS3Bucket),
		S3Region:                         envOrDefault(envS3Region, defaultS3Region),
		S3Secure:                         boolEnvOrDefault(envS3Secure, defaultS3Secure),
		MaxAttachmentBytes:               intEnvOrDefault(envMaxAttachmentBytes, defaultMaxAttachmentBytes),
		PrimaryThumbnailWarmLimit:        intEnvOrDefault(envPrimaryThumbnailWarmLimit, defaultPrimaryThumbnailWarmLimit),
		PrimaryThumbnailWarmConcurrency:  intEnvOrDefault(envPrimaryThumbnailWarmConcurrent, defaultPrimaryThumbnailWarmConcurrent),
		PrimaryThumbnailWarmTimeout:      durationEnvOrDefault(envPrimaryThumbnailWarmTimeout, defaultPrimaryThumbnailWarmTimeout),
		VoiceDevFakeEnabled:              boolEnvOrDefault(envVoiceDevFakeEnabled, defaultVoiceDevFakeEnabled),
		VoiceGoogleEnabled:               boolEnvOrDefault(envVoiceGoogleEnabled, defaultVoiceGoogleEnabled),
		RealtimeVoiceIdleTimeout:         durationEnvOrDefault(envRealtimeVoiceIdleTimeout, defaultRealtimeVoiceIdleTimeout),
		VoiceProviderHTTPTimeout:         durationEnvOrDefault(envVoiceProviderHTTPTimeout, defaultVoiceProviderHTTPTimeout),
		GoogleCloudProject:               os.Getenv(envGoogleCloudProject),
		GoogleCloudLocation:              envOrDefault(envGoogleCloudLocation, defaultGoogleCloudLocation),
		GoogleGeminiModel:                envOrDefault(envGoogleGeminiModel, defaultGoogleGeminiModel),
		GoogleTTSLanguageCode:            envOrDefault(envGoogleTTSLanguageCode, defaultGoogleTTSLanguageCode),
		GoogleTTSVoiceName:               envOrDefault(envGoogleTTSVoiceName, defaultGoogleTTSVoiceName),
		GoogleCredentialMode:             envOrDefault(envGoogleCredentialMode, defaultGoogleCredentialMode),
		GoogleAccessToken:                os.Getenv(envGoogleAccessToken),
		ProviderCredentialKeyID:          os.Getenv(envProviderCredentialKeyID),
		ProviderCredentialKey:            os.Getenv(envProviderCredentialKey),
		ImportJobTimeout:                 secondsEnvOrDefault(envImportJobTimeoutSeconds, defaultImportJobTimeout),
		ImportCredentialVacuumInterval:   secondsEnvOrDefault(envImportCredentialVacuumSeconds, defaultImportCredentialVacuumInterval),
	}
}

func oidcMobileScopes() []string {
	scopes := stringListEnv(envOIDCMobileScopes)
	if len(scopes) > 0 {
		return scopes
	}
	return []string{"openid", "email", "profile", "offline_access"}
}

func oidcClientIDs() []string {
	clientIDs := stringListEnv(envOIDCClientIDs)
	clientIDs = prependUniqueClientID(clientIDs, strings.TrimSpace(os.Getenv(envOIDCClientID)))
	clientIDs = appendUniqueClientID(clientIDs, strings.TrimSpace(os.Getenv(envOIDCMobileClientID)))
	return clientIDs
}

func prependUniqueClientID(clientIDs []string, clientID string) []string {
	if clientID == "" || containsString(clientIDs, clientID) {
		return clientIDs
	}
	return append([]string{clientID}, clientIDs...)
}

func appendUniqueClientID(clientIDs []string, clientID string) []string {
	if clientID == "" || containsString(clientIDs, clientID) {
		return clientIDs
	}
	return append(clientIDs, clientID)
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
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

func int64EnvOrDefault(name string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
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

func secondsEnvOrDefault(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return time.Duration(parsed) * time.Second
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
