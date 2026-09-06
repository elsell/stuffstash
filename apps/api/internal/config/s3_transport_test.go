package config

import "testing"

func TestPublicS3TLSDefaultsToInternalAndCanBeIndependent(t *testing.T) {
	t.Setenv("STUFF_STASH_S3_SECURE", "false")
	t.Setenv("STUFF_STASH_S3_PUBLIC_SECURE", "")
	cfg := Load()
	if cfg.S3PublicSecure {
		t.Fatal("public TLS must preserve existing internal default")
	}
	t.Setenv("STUFF_STASH_S3_PUBLIC_SECURE", "true")
	cfg = Load()
	if cfg.S3Secure || !cfg.S3PublicSecure {
		t.Fatal("public TLS not independently configured")
	}
}
