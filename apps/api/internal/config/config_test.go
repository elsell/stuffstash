package config

import (
	"testing"
	"time"
)

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv(envHTTPAddr, "")
	t.Setenv(envHTTPReadHeaderTimeout, "")
	t.Setenv(envHTTPReadTimeout, "")
	t.Setenv(envHTTPWriteTimeout, "")
	t.Setenv(envHTTPIdleTimeout, "")
	t.Setenv(envHTTPMaxJSONBodyBytes, "")
	t.Setenv(envHTTPRateLimitEnabled, "")
	t.Setenv(envHTTPRateLimitRequests, "")
	t.Setenv(envHTTPRateLimitWindow, "")
	t.Setenv(envHTTPRateLimitBurst, "")
	t.Setenv(envCORSAllowedOrigins, "")
	t.Setenv(envAuthMode, "")
	t.Setenv(envAuthzMode, "")
	t.Setenv(envRepositoryMode, "")
	t.Setenv(envDatabaseDSN, "")
	t.Setenv(envSpiceDBTLSEnabled, "")
	t.Setenv(envSpiceDBCAPath, "")
	t.Setenv(envSpiceDBBootstrapSchema, "")
	t.Setenv(envSpiceDBSchemaPath, "")
	t.Setenv(envAuthorizationOutboxLimit, "")
	t.Setenv(envAuthorizationOutboxEvery, "")
	t.Setenv(envAuthorizationOutboxLease, "")
	t.Setenv(envBlobDeletionOutboxLimit, "")
	t.Setenv(envBlobDeletionOutboxEvery, "")
	t.Setenv(envBlobDeletionOutboxLease, "")
	t.Setenv(envBlobDeletionOutboxMaxAttempts, "")
	t.Setenv(envInvitationTTL, "")
	t.Setenv(envInvitationPublicBaseURL, "")
	t.Setenv(envInvitationAllowInsecureLocalHTTP, "")
	t.Setenv(envDefaultPageLimit, "")
	t.Setenv(envMaxPageLimit, "")
	t.Setenv(envBlobStorageMode, "")
	t.Setenv(envBlobStoragePath, "")
	t.Setenv(envS3Endpoint, "")
	t.Setenv(envS3AccessKey, "")
	t.Setenv(envS3SecretKey, "")
	t.Setenv(envS3Bucket, "")
	t.Setenv(envS3Region, "")
	t.Setenv(envS3Secure, "")
	t.Setenv(envMaxAttachmentBytes, "")
	t.Setenv(envPrimaryThumbnailWarmLimit, "")
	t.Setenv(envPrimaryThumbnailWarmConcurrent, "")
	t.Setenv(envPrimaryThumbnailWarmTimeout, "")
	t.Setenv(envVoiceDevFakeEnabled, "")
	t.Setenv(envVoiceGoogleEnabled, "")
	t.Setenv(envRealtimeVoiceIdleTimeout, "")
	t.Setenv(envRealtimeVoiceToolCallTimeout, "")
	t.Setenv(envVoiceProviderHTTPTimeout, "")
	t.Setenv(envGoogleCloudProject, "")
	t.Setenv(envGoogleCloudLocation, "")
	t.Setenv(envGoogleGeminiModel, "")
	t.Setenv(envGoogleTTSLanguageCode, "")
	t.Setenv(envGoogleTTSVoiceName, "")
	t.Setenv(envGoogleCredentialMode, "")
	t.Setenv(envGoogleAccessToken, "")
	t.Setenv(envProviderCredentialKeyID, "")
	t.Setenv(envProviderCredentialKey, "")

	cfg := Load()

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("expected HTTP addr %q, got %q", defaultHTTPAddr, cfg.HTTPAddr)
	}
	if cfg.HTTPReadHeaderTimeout != defaultHTTPReadHeader || cfg.HTTPReadTimeout != defaultHTTPRead || cfg.HTTPWriteTimeout != defaultHTTPWrite || cfg.HTTPIdleTimeout != defaultHTTPIdle {
		t.Fatalf("unexpected HTTP timeout defaults: %+v", cfg)
	}
	if cfg.HTTPMaxJSONBodyBytes != defaultHTTPMaxJSONBodyBytes {
		t.Fatalf("expected default max JSON body bytes %d, got %d", defaultHTTPMaxJSONBodyBytes, cfg.HTTPMaxJSONBodyBytes)
	}
	if !cfg.HTTPRateLimitEnabled || cfg.HTTPRateLimitRequests != defaultHTTPRateLimitRequests || cfg.HTTPRateLimitWindow != defaultHTTPRateLimitWindow || cfg.HTTPRateLimitBurst != defaultHTTPRateLimitBurst {
		t.Fatalf("unexpected HTTP rate limit defaults: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 0 {
		t.Fatalf("expected no CORS allowed origins by default, got %+v", cfg.CORSAllowedOrigins)
	}
	if cfg.AuthMode != defaultAuthMode {
		t.Fatalf("expected auth mode %q, got %q", defaultAuthMode, cfg.AuthMode)
	}
	if cfg.AuthzMode != defaultAuthzMode {
		t.Fatalf("expected authz mode %q, got %q", defaultAuthzMode, cfg.AuthzMode)
	}
	if cfg.RepositoryMode != defaultRepositoryMode {
		t.Fatalf("expected repository mode %q, got %q", defaultRepositoryMode, cfg.RepositoryMode)
	}
	if cfg.DatabaseDSN != "" {
		t.Fatalf("expected empty database dsn, got %q", cfg.DatabaseDSN)
	}
	if !cfg.SpiceDBTLSEnabled {
		t.Fatalf("expected SpiceDB TLS to default enabled")
	}
	if cfg.SpiceDBCAPath != "" {
		t.Fatalf("expected empty SpiceDB CA path, got %q", cfg.SpiceDBCAPath)
	}
	if cfg.SpiceDBBootstrapSchema {
		t.Fatalf("expected SpiceDB schema bootstrap to default disabled")
	}
	if cfg.SpiceDBSchemaPath != defaultSpiceDBSchemaPath {
		t.Fatalf("expected schema path %q, got %q", defaultSpiceDBSchemaPath, cfg.SpiceDBSchemaPath)
	}
	if cfg.AuthorizationOutboxDrainLimit != defaultAuthorizationLimit {
		t.Fatalf("expected authorization outbox limit %d, got %d", defaultAuthorizationLimit, cfg.AuthorizationOutboxDrainLimit)
	}
	if cfg.AuthorizationOutboxDrainInterval != defaultAuthorizationEvery {
		t.Fatalf("expected authorization outbox interval %s, got %s", defaultAuthorizationEvery, cfg.AuthorizationOutboxDrainInterval)
	}
	if cfg.AuthorizationOutboxClaimLease != defaultAuthorizationLease {
		t.Fatalf("expected authorization outbox claim lease %s, got %s", defaultAuthorizationLease, cfg.AuthorizationOutboxClaimLease)
	}
	if cfg.BlobDeletionOutboxDrainLimit != defaultBlobDeletionLimit {
		t.Fatalf("expected blob deletion outbox limit %d, got %d", defaultBlobDeletionLimit, cfg.BlobDeletionOutboxDrainLimit)
	}
	if cfg.BlobDeletionOutboxDrainInterval != defaultBlobDeletionEvery {
		t.Fatalf("expected blob deletion outbox interval %s, got %s", defaultBlobDeletionEvery, cfg.BlobDeletionOutboxDrainInterval)
	}
	if cfg.BlobDeletionOutboxClaimLease != defaultBlobDeletionLease {
		t.Fatalf("expected blob deletion outbox claim lease %s, got %s", defaultBlobDeletionLease, cfg.BlobDeletionOutboxClaimLease)
	}
	if cfg.BlobDeletionOutboxMaxAttempts != defaultBlobDeletionMaxAttempts {
		t.Fatalf("expected blob deletion outbox max attempts %d, got %d", defaultBlobDeletionMaxAttempts, cfg.BlobDeletionOutboxMaxAttempts)
	}
	if cfg.InvitationTTL != defaultInvitationTTL {
		t.Fatalf("expected invitation TTL %s, got %s", defaultInvitationTTL, cfg.InvitationTTL)
	}
	if cfg.InvitationPublicBaseURL != defaultInvitationPublicBaseURL {
		t.Fatalf("expected default invitation public base URL %q, got %q", defaultInvitationPublicBaseURL, cfg.InvitationPublicBaseURL)
	}
	if cfg.InvitationAllowInsecureLocalHTTP {
		t.Fatalf("expected insecure local invitation HTTP to default disabled")
	}
	if cfg.DefaultPageLimit != defaultDefaultPageLimit {
		t.Fatalf("expected default page limit %d, got %d", defaultDefaultPageLimit, cfg.DefaultPageLimit)
	}
	if cfg.MaxPageLimit != defaultMaxPageLimit {
		t.Fatalf("expected max page limit %d, got %d", defaultMaxPageLimit, cfg.MaxPageLimit)
	}
	if cfg.BlobStorageMode != defaultBlobStorageMode {
		t.Fatalf("expected blob storage mode %q, got %q", defaultBlobStorageMode, cfg.BlobStorageMode)
	}
	if cfg.BlobStoragePath != defaultBlobStoragePath {
		t.Fatalf("expected blob storage path %q, got %q", defaultBlobStoragePath, cfg.BlobStoragePath)
	}
	if cfg.S3Endpoint != "" || cfg.S3PublicEndpoint != "" || cfg.S3AccessKey != "" || cfg.S3SecretKey != "" || cfg.S3Bucket != "" {
		t.Fatalf("expected empty S3 connection fields, got %+v", cfg)
	}
	if cfg.S3Region != defaultS3Region {
		t.Fatalf("expected S3 region %q, got %q", defaultS3Region, cfg.S3Region)
	}
	if !cfg.S3Secure {
		t.Fatalf("expected S3 secure to default enabled")
	}
	if cfg.MaxAttachmentBytes != defaultMaxAttachmentBytes {
		t.Fatalf("expected max attachment bytes %d, got %d", defaultMaxAttachmentBytes, cfg.MaxAttachmentBytes)
	}
	if cfg.PrimaryThumbnailWarmLimit != defaultPrimaryThumbnailWarmLimit || cfg.PrimaryThumbnailWarmConcurrency != defaultPrimaryThumbnailWarmConcurrent || cfg.PrimaryThumbnailWarmTimeout != defaultPrimaryThumbnailWarmTimeout {
		t.Fatalf("unexpected primary thumbnail warm defaults: %+v", cfg)
	}
	if cfg.VoiceDevFakeEnabled {
		t.Fatalf("expected voice dev fake providers disabled by default")
	}
	if cfg.VoiceGoogleEnabled {
		t.Fatalf("expected Google voice providers disabled by default")
	}
	if cfg.RealtimeVoiceIdleTimeout != defaultRealtimeVoiceIdleTimeout {
		t.Fatalf("expected realtime voice idle timeout %s, got %s", defaultRealtimeVoiceIdleTimeout, cfg.RealtimeVoiceIdleTimeout)
	}
	if cfg.RealtimeVoiceToolCallTimeout != defaultRealtimeVoiceToolCallTimeout {
		t.Fatalf("expected realtime voice tool-call timeout %s, got %s", defaultRealtimeVoiceToolCallTimeout, cfg.RealtimeVoiceToolCallTimeout)
	}
	if cfg.VoiceProviderHTTPTimeout != defaultVoiceProviderHTTPTimeout {
		t.Fatalf("expected voice provider HTTP timeout %s, got %s", defaultVoiceProviderHTTPTimeout, cfg.VoiceProviderHTTPTimeout)
	}
	if cfg.GoogleCloudProject != "" || cfg.GoogleCloudLocation != defaultGoogleCloudLocation || cfg.GoogleGeminiModel != defaultGoogleGeminiModel || cfg.GoogleTTSLanguageCode != defaultGoogleTTSLanguageCode || cfg.GoogleTTSVoiceName != defaultGoogleTTSVoiceName {
		t.Fatalf("unexpected Google voice defaults: %+v", cfg)
	}
	if cfg.GoogleCredentialMode != defaultGoogleCredentialMode {
		t.Fatalf("expected Google credential mode %q, got %q", defaultGoogleCredentialMode, cfg.GoogleCredentialMode)
	}
	if cfg.GoogleAccessToken != "" {
		t.Fatalf("expected empty Google access token by default")
	}
	if cfg.ProviderCredentialKeyID != "" || cfg.ProviderCredentialKey != "" {
		t.Fatalf("expected empty provider credential encryption config by default")
	}
	if cfg.ImportJobTimeout != 15*time.Minute {
		t.Fatalf("expected import job timeout default 15m, got %s", cfg.ImportJobTimeout)
	}
	if cfg.ImportCredentialVacuumInterval != time.Minute {
		t.Fatalf("expected import credential vacuum interval default 1m, got %s", cfg.ImportCredentialVacuumInterval)
	}
}

func TestLoadReadsAuthAndSpiceDBConfiguration(t *testing.T) {
	t.Setenv(envAuthMode, "oidc")
	t.Setenv(envHTTPReadHeaderTimeout, "2s")
	t.Setenv(envHTTPReadTimeout, "3s")
	t.Setenv(envHTTPWriteTimeout, "4s")
	t.Setenv(envHTTPIdleTimeout, "5s")
	t.Setenv(envHTTPMaxJSONBodyBytes, "2048")
	t.Setenv(envHTTPRateLimitEnabled, "false")
	t.Setenv(envHTTPRateLimitRequests, "12")
	t.Setenv(envHTTPRateLimitWindow, "30s")
	t.Setenv(envHTTPRateLimitBurst, "4")
	t.Setenv(envCORSAllowedOrigins, "http://localhost:5173, https://stuffstash.online, http://localhost:5173")
	t.Setenv(envAuthzMode, "spicedb")
	t.Setenv(envOIDCIssuer, "https://accounts.google.com")
	t.Setenv(envOIDCClientID, "client-id")
	t.Setenv(envOIDCClientIDs, "web-client-id, mobile-client-id, client-id")
	t.Setenv(envOIDCMobileClientID, "mobile-client-id")
	t.Setenv(envOIDCMobileRedirectURI, "stuffstash://auth/callback")
	t.Setenv(envOIDCMobileScopes, "openid,email,profile,offline_access")
	t.Setenv(envRepositoryMode, "postgres")
	t.Setenv(envDatabaseDSN, "postgres://stuffstash:stuffstash-local@postgres:5432/stuffstash?sslmode=disable")
	t.Setenv(envSpiceDBEndpoint, "spicedb:50051")
	t.Setenv(envSpiceDBPresharedKey, "local-key")
	t.Setenv(envSpiceDBTLSEnabled, "false")
	t.Setenv(envSpiceDBCAPath, "/var/run/stuffstash/spicedb-ca/ca.crt")
	t.Setenv(envSpiceDBBootstrapSchema, "true")
	t.Setenv(envSpiceDBSchemaPath, "custom/schema.zed")
	t.Setenv(envAuthorizationOutboxLimit, "7")
	t.Setenv(envAuthorizationOutboxEvery, "250ms")
	t.Setenv(envAuthorizationOutboxLease, "45s")
	t.Setenv(envBlobDeletionOutboxLimit, "9")
	t.Setenv(envBlobDeletionOutboxEvery, "750ms")
	t.Setenv(envBlobDeletionOutboxLease, "55s")
	t.Setenv(envBlobDeletionOutboxMaxAttempts, "3")
	t.Setenv(envInvitationTTL, "2h")
	t.Setenv(envInvitationPublicBaseURL, "https://stash.example.test/invitations/accept")
	t.Setenv(envInvitationAllowInsecureLocalHTTP, "true")
	t.Setenv(envDefaultPageLimit, "13")
	t.Setenv(envMaxPageLimit, "27")
	t.Setenv(envBlobStorageMode, "s3")
	t.Setenv(envBlobStoragePath, "/data/blobs")
	t.Setenv(envS3Endpoint, "garage:3900")
	t.Setenv(envS3PublicEndpoint, "localhost:3900")
	t.Setenv(envS3AccessKey, "access")
	t.Setenv(envS3SecretKey, "secret")
	t.Setenv(envS3Bucket, "stuffstash")
	t.Setenv(envS3Region, "local")
	t.Setenv(envS3Secure, "false")
	t.Setenv(envMaxAttachmentBytes, "12345")
	t.Setenv(envPrimaryThumbnailWarmLimit, "8")
	t.Setenv(envPrimaryThumbnailWarmConcurrent, "2")
	t.Setenv(envPrimaryThumbnailWarmTimeout, "3s")
	t.Setenv(envVoiceDevFakeEnabled, "true")
	t.Setenv(envVoiceGoogleEnabled, "true")
	t.Setenv(envRealtimeVoiceIdleTimeout, "12s")
	t.Setenv(envRealtimeVoiceToolCallTimeout, "4s")
	t.Setenv(envVoiceProviderHTTPTimeout, "75s")
	t.Setenv(envGoogleCloudProject, "pianotechpros")
	t.Setenv(envGoogleCloudLocation, "us-east5")
	t.Setenv(envGoogleGeminiModel, "gemini-test-model")
	t.Setenv(envGoogleTTSLanguageCode, "en-GB")
	t.Setenv(envGoogleTTSVoiceName, "en-GB-Neural2-A")
	t.Setenv(envGoogleCredentialMode, GoogleCredentialModeAccessToken)
	t.Setenv(envGoogleAccessToken, "ya29.test")
	t.Setenv(envProviderCredentialKeyID, "local-key")
	t.Setenv(envProviderCredentialKey, "base64-key")
	t.Setenv(envImportJobTimeoutSeconds, "120")
	t.Setenv(envImportCredentialVacuumSeconds, "30")

	cfg := Load()

	if cfg.AuthMode != "oidc" {
		t.Fatalf("expected auth mode oidc, got %q", cfg.AuthMode)
	}
	if cfg.HTTPReadHeaderTimeout.String() != "2s" || cfg.HTTPReadTimeout.String() != "3s" || cfg.HTTPWriteTimeout.String() != "4s" || cfg.HTTPIdleTimeout.String() != "5s" {
		t.Fatalf("unexpected HTTP timeout config: %+v", cfg)
	}
	if cfg.HTTPMaxJSONBodyBytes != 2048 {
		t.Fatalf("expected max JSON body bytes 2048, got %d", cfg.HTTPMaxJSONBodyBytes)
	}
	if cfg.HTTPRateLimitEnabled || cfg.HTTPRateLimitRequests != 12 || cfg.HTTPRateLimitWindow.String() != "30s" || cfg.HTTPRateLimitBurst != 4 {
		t.Fatalf("unexpected HTTP rate limit config: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "http://localhost:5173" || cfg.CORSAllowedOrigins[1] != "https://stuffstash.online" {
		t.Fatalf("unexpected CORS allowed origins: %+v", cfg.CORSAllowedOrigins)
	}
	if cfg.AuthzMode != "spicedb" {
		t.Fatalf("expected authz mode spicedb, got %q", cfg.AuthzMode)
	}
	if cfg.OIDCIssuer != "https://accounts.google.com" || cfg.OIDCClientID != "client-id" {
		t.Fatalf("unexpected OIDC config: %+v", cfg)
	}
	if len(cfg.OIDCClientIDs) != 3 || cfg.OIDCClientIDs[0] != "web-client-id" || cfg.OIDCClientIDs[1] != "mobile-client-id" || cfg.OIDCClientIDs[2] != "client-id" {
		t.Fatalf("unexpected OIDC client IDs: %+v", cfg.OIDCClientIDs)
	}
	if cfg.OIDCMobileClientID != "mobile-client-id" || cfg.OIDCMobileRedirectURI != "stuffstash://auth/callback" {
		t.Fatalf("unexpected mobile OIDC config: %+v", cfg)
	}
	if len(cfg.OIDCMobileScopes) != 4 || cfg.OIDCMobileScopes[0] != "openid" || cfg.OIDCMobileScopes[3] != "offline_access" {
		t.Fatalf("unexpected mobile OIDC scopes: %+v", cfg.OIDCMobileScopes)
	}
	if cfg.RepositoryMode != "postgres" || cfg.DatabaseDSN == "" {
		t.Fatalf("unexpected repository config: %+v", cfg)
	}
	if cfg.SpiceDBEndpoint != "spicedb:50051" || cfg.SpiceDBPresharedKey != "local-key" {
		t.Fatalf("unexpected SpiceDB config: %+v", cfg)
	}
	if cfg.SpiceDBTLSEnabled {
		t.Fatalf("expected SpiceDB TLS disabled")
	}
	if cfg.SpiceDBCAPath != "/var/run/stuffstash/spicedb-ca/ca.crt" {
		t.Fatalf("expected custom SpiceDB CA path, got %q", cfg.SpiceDBCAPath)
	}
	if !cfg.SpiceDBBootstrapSchema {
		t.Fatalf("expected SpiceDB schema bootstrap enabled")
	}
	if cfg.SpiceDBSchemaPath != "custom/schema.zed" {
		t.Fatalf("expected custom schema path, got %q", cfg.SpiceDBSchemaPath)
	}
	if cfg.AuthorizationOutboxDrainLimit != 7 {
		t.Fatalf("expected authorization outbox drain limit 7, got %d", cfg.AuthorizationOutboxDrainLimit)
	}
	if cfg.AuthorizationOutboxDrainInterval.String() != "250ms" {
		t.Fatalf("expected authorization outbox drain interval 250ms, got %s", cfg.AuthorizationOutboxDrainInterval)
	}
	if cfg.AuthorizationOutboxClaimLease.String() != "45s" {
		t.Fatalf("expected authorization outbox claim lease 45s, got %s", cfg.AuthorizationOutboxClaimLease)
	}
	if cfg.BlobDeletionOutboxDrainLimit != 9 {
		t.Fatalf("expected blob deletion outbox drain limit 9, got %d", cfg.BlobDeletionOutboxDrainLimit)
	}
	if cfg.BlobDeletionOutboxDrainInterval.String() != "750ms" {
		t.Fatalf("expected blob deletion outbox drain interval 750ms, got %s", cfg.BlobDeletionOutboxDrainInterval)
	}
	if cfg.BlobDeletionOutboxClaimLease.String() != "55s" {
		t.Fatalf("expected blob deletion outbox claim lease 55s, got %s", cfg.BlobDeletionOutboxClaimLease)
	}
	if cfg.BlobDeletionOutboxMaxAttempts != 3 {
		t.Fatalf("expected blob deletion outbox max attempts 3, got %d", cfg.BlobDeletionOutboxMaxAttempts)
	}
	if cfg.InvitationTTL.String() != "2h0m0s" {
		t.Fatalf("expected invitation TTL 2h, got %s", cfg.InvitationTTL)
	}
	if cfg.InvitationPublicBaseURL != "https://stash.example.test/invitations/accept" {
		t.Fatalf("unexpected invitation public base URL %q", cfg.InvitationPublicBaseURL)
	}
	if !cfg.InvitationAllowInsecureLocalHTTP {
		t.Fatalf("expected insecure local invitation HTTP override enabled")
	}
	if cfg.DefaultPageLimit != 13 || cfg.MaxPageLimit != 27 {
		t.Fatalf("unexpected page limits: %+v", cfg)
	}
	if cfg.BlobStorageMode != "s3" {
		t.Fatalf("expected blob storage mode s3, got %q", cfg.BlobStorageMode)
	}
	if cfg.BlobStoragePath != "/data/blobs" {
		t.Fatalf("expected custom blob storage path, got %q", cfg.BlobStoragePath)
	}
	if cfg.S3Endpoint != "garage:3900" || cfg.S3PublicEndpoint != "localhost:3900" || cfg.S3AccessKey != "access" || cfg.S3SecretKey != "secret" || cfg.S3Bucket != "stuffstash" {
		t.Fatalf("unexpected S3 config: %+v", cfg)
	}
	if cfg.S3Region != "local" || cfg.S3Secure {
		t.Fatalf("unexpected S3 region or secure setting: %+v", cfg)
	}
	if cfg.MaxAttachmentBytes != 12345 {
		t.Fatalf("expected max attachment bytes 12345, got %d", cfg.MaxAttachmentBytes)
	}
	if cfg.PrimaryThumbnailWarmLimit != 8 || cfg.PrimaryThumbnailWarmConcurrency != 2 || cfg.PrimaryThumbnailWarmTimeout.String() != "3s" {
		t.Fatalf("unexpected primary thumbnail warm config: %+v", cfg)
	}
	if !cfg.VoiceDevFakeEnabled {
		t.Fatalf("expected voice dev fake providers enabled")
	}
	if cfg.RealtimeVoiceIdleTimeout.String() != "12s" {
		t.Fatalf("expected realtime voice idle timeout 12s, got %s", cfg.RealtimeVoiceIdleTimeout)
	}
	if cfg.RealtimeVoiceToolCallTimeout.String() != "4s" {
		t.Fatalf("expected realtime voice tool-call timeout 4s, got %s", cfg.RealtimeVoiceToolCallTimeout)
	}
	if cfg.VoiceProviderHTTPTimeout.String() != "1m15s" {
		t.Fatalf("expected voice provider HTTP timeout 75s, got %s", cfg.VoiceProviderHTTPTimeout)
	}
	if !cfg.VoiceGoogleEnabled || cfg.GoogleCloudProject != "pianotechpros" || cfg.GoogleCloudLocation != "us-east5" || cfg.GoogleGeminiModel != "gemini-test-model" || cfg.GoogleTTSLanguageCode != "en-GB" || cfg.GoogleTTSVoiceName != "en-GB-Neural2-A" {
		t.Fatalf("unexpected Google voice config: %+v", cfg)
	}
	if cfg.GoogleCredentialMode != GoogleCredentialModeAccessToken {
		t.Fatalf("expected Google access token credential mode, got %q", cfg.GoogleCredentialMode)
	}
	if cfg.GoogleAccessToken != "ya29.test" {
		t.Fatalf("expected Google access token config")
	}
	if cfg.ProviderCredentialKeyID != "local-key" || cfg.ProviderCredentialKey != "base64-key" {
		t.Fatalf("unexpected provider credential encryption config: %+v", cfg)
	}
	if cfg.ImportJobTimeout.String() != "2m0s" {
		t.Fatalf("expected import job timeout 120s, got %s", cfg.ImportJobTimeout)
	}
	if cfg.ImportCredentialVacuumInterval.String() != "30s" {
		t.Fatalf("expected import credential vacuum interval 30s, got %s", cfg.ImportCredentialVacuumInterval)
	}
}

func TestLoadAddsMobileClientIDToAcceptedOIDCAudiences(t *testing.T) {
	t.Setenv(envOIDCClientID, "api-client-id")
	t.Setenv(envOIDCClientIDs, "web-client-id")
	t.Setenv(envOIDCMobileClientID, "mobile-client-id")

	cfg := Load()

	if len(cfg.OIDCClientIDs) != 3 {
		t.Fatalf("expected 3 accepted client IDs, got %+v", cfg.OIDCClientIDs)
	}
	if cfg.OIDCClientIDs[0] != "api-client-id" || cfg.OIDCClientIDs[1] != "web-client-id" || cfg.OIDCClientIDs[2] != "mobile-client-id" {
		t.Fatalf("unexpected accepted client IDs: %+v", cfg.OIDCClientIDs)
	}
}
