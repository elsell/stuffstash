package httpserver

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func seedLiveBabyClothes(t *testing.T, ctx context.Context, application app.App) liveAudioFixture {
	t.Helper()
	principal := identity.Principal{ID: "user-1"}
	tags := []string{}
	for _, name := range []string{"baby", "clothes"} {
		tag, err := application.CreateAssetTag(ctx, app.CreateAssetTagInput{Principal: principal, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Key: name, DisplayName: name, Color: "#2f80ed"})
		if err != nil {
			t.Fatal(err)
		}
		tags = append(tags, tag.ID.String())
	}
	create := func(kind, title, parent string, tagIDs []string) string {
		t.Helper()
		r, err := application.CreateAssetWithOperation(ctx, app.CreateAssetInput{Principal: principal, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: kind, Title: title, ParentAssetID: parent, TagIDs: tagIDs})
		if err != nil {
			t.Fatal(err)
		}
		return r.Asset.ID.String()
	}
	closet := create("location", "Hall Closet", "", nil)
	bin58 := create("container", "Bin 58", closet, nil)
	bin106 := create("container", "Bin 106", closet, nil)
	first := create("item", "3–6 months clothes", bin58, tags)
	second := create("item", "6–9 months", bin106, tags)
	exact := create("item", "Baby clothes", bin58, nil)
	create("item", "Adult winter jacket", closet, nil)

	return liveAudioFixture{expectedIDs: []string{first, second, exact}, spokenLocations: []string{"Bin 58", "Bin 106"}}
}

func seedLiveChemicals(t *testing.T, ctx context.Context, application app.App) liveAudioFixture {
	t.Helper()
	principal := identity.Principal{ID: "user-1"}
	tag, err := application.CreateAssetTag(ctx, app.CreateAssetTagInput{Principal: principal, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Key: "chemicals", DisplayName: "Chemicals", Color: "#2f80ed"})
	if err != nil {
		t.Fatal(err)
	}
	create := func(kind, title, parent string, tags []string) string {
		t.Helper()
		result, err := application.CreateAssetWithOperation(ctx, app.CreateAssetInput{Principal: principal, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: kind, Title: title, ParentAssetID: parent, TagIDs: tags})
		if err != nil {
			t.Fatal(err)
		}
		return result.Asset.ID.String()
	}
	garage := create("location", "Garage", "", nil)
	cabinet := create("container", "Cabinet A", garage, nil)
	tags := []string{tag.ID.String()}
	acetone := create("item", "Acetone", cabinet, tags)
	carbonate := create("item", "Sodium carbonate", cabinet, tags)
	book := create("item", "Household chemicals reference book", garage, nil)
	gloves := create("item", "Chemicals protective gloves", garage, nil)
	jacket := create("item", "Winter jacket", garage, nil)
	return liveAudioFixture{expectedIDs: []string{acetone, carbonate}, excludedIDs: []string{book, gloves, jacket}}
}
