package homebox

import (
	"context"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestLegacyImporterReadsLiveHomebox(t *testing.T) {
	if os.Getenv("STUFFSTASH_HOMEBOX_LIVE") != "1" {
		t.Skip("set STUFFSTASH_HOMEBOX_LIVE=1 to run against a live Homebox instance")
	}
	baseURL := os.Getenv("STUFFSTASH_HOMEBOX_URL")
	username := os.Getenv("STUFFSTASH_HOMEBOX_USERNAME")
	password := os.Getenv("STUFFSTASH_HOMEBOX_PASSWORD")
	if baseURL == "" || username == "" || password == "" {
		t.Fatal("STUFFSTASH_HOMEBOX_URL, STUFFSTASH_HOMEBOX_USERNAME, and STUFFSTASH_HOMEBOX_PASSWORD are required")
	}

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType:          importplan.SourceLegacyHomebox,
		BaseURL:             baseURL,
		Username:            username,
		Password:            password,
		IncludeImages:       os.Getenv("STUFFSTASH_HOMEBOX_INCLUDE_IMAGES") == "1",
		AllowInsecureTLS:    os.Getenv("STUFFSTASH_HOMEBOX_ALLOW_INSECURE_TLS") == "1",
		AllowPrivateNetwork: os.Getenv("STUFFSTASH_HOMEBOX_ALLOW_PRIVATE_NETWORK") == "1",
	})
	if err != nil {
		t.Fatalf("read live Homebox import plan: %v", err)
	}
	counts := plan.Counts()
	t.Logf("read Homebox %s: locations=%d assets=%d attachments=%d warnings=%d errors=%d", plan.Source.Version, counts.Locations, counts.Assets, counts.Attachments, counts.Warnings, counts.Errors)
	if counts.Assets == 0 {
		t.Fatalf("expected live Homebox assets, got counts %+v", counts)
	}
	if plan.Source.Version == "" {
		t.Fatalf("expected Homebox version in source summary, got %+v", plan.Source)
	}
	if os.Getenv("STUFFSTASH_HOMEBOX_INCLUDE_IMAGES") == "1" && counts.Attachments == 0 {
		t.Fatalf("expected live Homebox attachments when image import is enabled, got counts %+v", counts)
	}
}
