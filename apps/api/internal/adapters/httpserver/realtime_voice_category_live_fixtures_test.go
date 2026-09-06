package httpserver

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func seedLiveChemicalsWithoutLiteralMatches(t *testing.T, ctx context.Context, application app.App) liveAudioFixture {
	t.Helper()
	principal := identity.Principal{ID: "user-1"}
	makeTag := func(name string) string {
		t.Helper()
		tag, err := application.CreateAssetTag(ctx, app.CreateAssetTagInput{Principal: principal, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Key: name, DisplayName: name})
		if err != nil {
			t.Fatal(err)
		}
		return tag.ID.String()
	}
	create := func(kind, title, parent string, tags []string) string {
		t.Helper()
		result, err := application.CreateAssetWithOperation(ctx, app.CreateAssetInput{Principal: principal, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: kind, Title: title, ParentAssetID: parent, TagIDs: tags})
		if err != nil {
			t.Fatal(err)
		}
		return result.Asset.ID.String()
	}
	cleaning, adhesives := makeTag("cleaning"), makeTag("adhesives")
	garage := create("location", "Garage", "", nil)
	cabinet := create("container", "Cabinet A", garage, nil)
	acetone := create("item", "Acetone", cabinet, []string{cleaning})
	sealant := create("item", "Silicone sealant", cabinet, []string{adhesives})
	brush := create("item", "Cleaning brush", garage, []string{cleaning})
	caps := create("item", "Caulk caps", garage, []string{adhesives})
	return liveAudioFixture{expectedIDs: []string{acetone, sealant}, excludedIDs: []string{brush, caps}, artifactLocations: map[string]string{acetone: "Cabinet A", sealant: "Cabinet A"}}
}
