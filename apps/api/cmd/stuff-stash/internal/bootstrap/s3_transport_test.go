package bootstrap

import (
	"github.com/stuffstash/stuff-stash/internal/config"
	"testing"
)

func TestBlobStorageRejectsInvalidPublicTLS(t *testing.T) {
	t.Setenv("STUFF_STASH_BLOB_STORAGE_MODE", "s3")
	t.Setenv("STUFF_STASH_S3_SECURE", "false")
	t.Setenv("STUFF_STASH_S3_PUBLIC_SECURE", "treu")
	cfg := config.Load()
	cfg.BlobStorageMode = "s3"
	if _, _, err := buildBlobStorage(cfg); err == nil || err.Error() != "invalid STUFF_STASH_S3_PUBLIC_SECURE setting" {
		t.Fatal("invalid TLS configuration must fail before creating S3 storage")
	}
}
