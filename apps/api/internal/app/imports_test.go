package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
)

func TestSortedImportAssetsOrdersDeepParentsBeforeChildren(t *testing.T) {
	items := []importplan.Asset{
		{SourceID: "location:bin", ParentSourceID: "location:shelf", Kind: "location", Title: "Bin"},
		{SourceID: "location:garage", Kind: "location", Title: "Garage"},
		{SourceID: "item:drill", ParentSourceID: "location:bin", Kind: "item", Title: "Drill"},
		{SourceID: "location:shelf", ParentSourceID: "location:garage", Kind: "location", Title: "Shelf"},
	}

	got := sortedImportAssets(items, "location")

	if len(got) != 3 {
		t.Fatalf("expected three locations, got %d", len(got))
	}
	for index, sourceID := range []string{"location:garage", "location:shelf", "location:bin"} {
		if got[index].SourceID != sourceID {
			t.Fatalf("expected source %q at index %d, got %q", sourceID, index, got[index].SourceID)
		}
	}
}
