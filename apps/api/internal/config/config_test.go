package config

import "testing"

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv(envHTTPAddr, "")
	t.Setenv(envAuthMode, "")
	t.Setenv(envAuthzMode, "")
	t.Setenv(envRepositoryMode, "")
	t.Setenv(envDatabaseDSN, "")
	t.Setenv(envSpiceDBTLSEnabled, "")
	t.Setenv(envSpiceDBBootstrapSchema, "")
	t.Setenv(envSpiceDBSchemaPath, "")
	t.Setenv(envAuthorizationOutboxLimit, "")
	t.Setenv(envAuthorizationOutboxEvery, "")
	t.Setenv(envAuthorizationOutboxLease, "")
	t.Setenv(envDefaultPageLimit, "")
	t.Setenv(envMaxPageLimit, "")

	cfg := Load()

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("expected HTTP addr %q, got %q", defaultHTTPAddr, cfg.HTTPAddr)
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
	if cfg.DefaultPageLimit != defaultDefaultPageLimit {
		t.Fatalf("expected default page limit %d, got %d", defaultDefaultPageLimit, cfg.DefaultPageLimit)
	}
	if cfg.MaxPageLimit != defaultMaxPageLimit {
		t.Fatalf("expected max page limit %d, got %d", defaultMaxPageLimit, cfg.MaxPageLimit)
	}
}

func TestLoadReadsAuthAndSpiceDBConfiguration(t *testing.T) {
	t.Setenv(envAuthMode, "oidc")
	t.Setenv(envAuthzMode, "spicedb")
	t.Setenv(envOIDCIssuer, "https://accounts.google.com")
	t.Setenv(envOIDCClientID, "client-id")
	t.Setenv(envRepositoryMode, "postgres")
	t.Setenv(envDatabaseDSN, "postgres://stuffstash:stuffstash-local@postgres:5432/stuffstash?sslmode=disable")
	t.Setenv(envSpiceDBEndpoint, "spicedb:50051")
	t.Setenv(envSpiceDBPresharedKey, "local-key")
	t.Setenv(envSpiceDBTLSEnabled, "false")
	t.Setenv(envSpiceDBBootstrapSchema, "true")
	t.Setenv(envSpiceDBSchemaPath, "custom/schema.zed")
	t.Setenv(envAuthorizationOutboxLimit, "7")
	t.Setenv(envAuthorizationOutboxEvery, "250ms")
	t.Setenv(envAuthorizationOutboxLease, "45s")
	t.Setenv(envDefaultPageLimit, "13")
	t.Setenv(envMaxPageLimit, "27")

	cfg := Load()

	if cfg.AuthMode != "oidc" {
		t.Fatalf("expected auth mode oidc, got %q", cfg.AuthMode)
	}
	if cfg.AuthzMode != "spicedb" {
		t.Fatalf("expected authz mode spicedb, got %q", cfg.AuthzMode)
	}
	if cfg.OIDCIssuer != "https://accounts.google.com" || cfg.OIDCClientID != "client-id" {
		t.Fatalf("unexpected OIDC config: %+v", cfg)
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
	if cfg.DefaultPageLimit != 13 || cfg.MaxPageLimit != 27 {
		t.Fatalf("unexpected page limits: %+v", cfg)
	}
}
