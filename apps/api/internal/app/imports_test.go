package app

import (
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
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

func TestImportSourceInputErrorOnlySurfacesTypedUserErrors(t *testing.T) {
	raw := importSourceInputError(errors.New("password=secret token=abc internal=/tmp/source.json"))
	if !errors.Is(raw, ErrInvalidInput) {
		t.Fatalf("expected raw source error to remain invalid input, got %v", raw)
	}
	var rawDetail ImportSourceInvalidInputError
	if errors.As(raw, &rawDetail) {
		t.Fatalf("expected raw source error detail to stay hidden, got %q", rawDetail.Detail)
	}

	safe := importSourceInputError(ports.NewImportSourceUserError("Homebox URL resolves to a blocked address"))
	var safeDetail ImportSourceInvalidInputError
	if !errors.As(safe, &safeDetail) {
		t.Fatalf("expected typed source user error detail, got %v", safe)
	}
	if safeDetail.Detail != "Homebox URL resolves to a blocked address" {
		t.Fatalf("unexpected safe detail %q", safeDetail.Detail)
	}
	if !errors.Is(safe, ErrInvalidInput) {
		t.Fatalf("expected safe source error to remain invalid input, got %v", safe)
	}
}
