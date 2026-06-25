package config

import "testing"

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
	t.Setenv(envVoiceDevFakeEnabled, "")

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
	if cfg.S3Endpoint != "" || cfg.S3AccessKey != "" || cfg.S3SecretKey != "" || cfg.S3Bucket != "" {
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
	if cfg.VoiceDevFakeEnabled {
		t.Fatalf("expected voice dev fake providers disabled by default")
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
	t.Setenv(envDefaultPageLimit, "13")
	t.Setenv(envMaxPageLimit, "27")
	t.Setenv(envBlobStorageMode, "s3")
	t.Setenv(envBlobStoragePath, "/data/blobs")
	t.Setenv(envS3Endpoint, "localhost:3900")
	t.Setenv(envS3AccessKey, "access")
	t.Setenv(envS3SecretKey, "secret")
	t.Setenv(envS3Bucket, "stuffstash")
	t.Setenv(envS3Region, "local")
	t.Setenv(envS3Secure, "false")
	t.Setenv(envMaxAttachmentBytes, "12345")
	t.Setenv(envVoiceDevFakeEnabled, "true")

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
	if cfg.DefaultPageLimit != 13 || cfg.MaxPageLimit != 27 {
		t.Fatalf("unexpected page limits: %+v", cfg)
	}
	if cfg.BlobStorageMode != "s3" {
		t.Fatalf("expected blob storage mode s3, got %q", cfg.BlobStorageMode)
	}
	if cfg.BlobStoragePath != "/data/blobs" {
		t.Fatalf("expected custom blob storage path, got %q", cfg.BlobStoragePath)
	}
	if cfg.S3Endpoint != "localhost:3900" || cfg.S3AccessKey != "access" || cfg.S3SecretKey != "secret" || cfg.S3Bucket != "stuffstash" {
		t.Fatalf("unexpected S3 config: %+v", cfg)
	}
	if cfg.S3Region != "local" || cfg.S3Secure {
		t.Fatalf("unexpected S3 region or secure setting: %+v", cfg)
	}
	if cfg.MaxAttachmentBytes != 12345 {
		t.Fatalf("expected max attachment bytes 12345, got %d", cfg.MaxAttachmentBytes)
	}
	if !cfg.VoiceDevFakeEnabled {
		t.Fatalf("expected voice dev fake providers enabled")
	}
}
