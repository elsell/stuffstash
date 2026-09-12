package expiration

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func workspaceItem(t *testing.T, id, date string, enabled bool) WorkspaceItem {
	t.Helper()
	precision := expirationdate.Day
	if len(date) == 7 {
		precision = expirationdate.Month
	}
	parsed, err := expirationdate.ParseDate(date, precision)
	if err != nil {
		t.Fatal(err)
	}
	settings := notification.DefaultSettings("UTC")
	settings.Defaults = notification.ExpirationPreferences{AdvanceDays: 7}
	description, err := Describe(parsed, "type", enabled, settings, time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return WorkspaceItem{Asset: asset.Asset{ID: asset.ID(id), Expiration: parsed}, Expiration: description}
}
func TestWorkspaceGroupsAndPaginatesDatesWithoutLosingPrecision(t *testing.T) {
	items := []WorkspaceItem{workspaceItem(t, "future", "2027-01", true), workspaceItem(t, "old", "2026-08", true), workspaceItem(t, "recent", "2026-09-11", true), workspaceItem(t, "soon", "2026-09-18", true), workspaceItem(t, "disabled", "2026-09-13", false)}
	selector := NewWorkspaceSelection(WorkspaceAll, 2, "")
	for _, item := range items {
		selector.Add(item)
	}
	first := selector.Result()
	if first.Counts.All != 5 || first.Counts.Expired != 2 || first.Counts.Soon != 1 || len(first.Items) != 2 || first.Items[0].Asset.ID != "recent" || first.Items[1].Asset.ID != "old" || !first.HasMore {
		t.Fatalf("bad first page: %+v", first)
	}
	selector = NewWorkspaceSelection(WorkspaceAll, 10, WorkspacePosition(first.Items[1], WorkspaceAll))
	for _, item := range items {
		selector.Add(item)
	}
	next := selector.Result()
	if len(next.Items) != 3 || next.Items[0].Asset.ID != "disabled" || next.Items[2].Asset.Expiration.Precision() != expirationdate.Month || next.Counts.All != 5 {
		t.Fatalf("bad next page: %+v", next)
	}
}
func TestWorkspaceSoonAndExpiredExcludeDisabledTracking(t *testing.T) {
	for _, mode := range []WorkspaceMode{WorkspaceSoon, WorkspaceExpired} {
		selector := NewWorkspaceSelection(mode, 10, "")
		for _, date := range []string{"2026-09-11", "2026-09-18"} {
			selector.Add(workspaceItem(t, "disabled", date, false))
		}
		if len(selector.Result().Items) != 0 {
			t.Fatal("disabled tracking counted as attention")
		}
	}
}
func TestWorkspaceFiltersIntersectAndKeepMonthEndSemantics(t *testing.T) {
	item := workspaceItem(t, "item", "2026-09", true)
	item.Asset.Title, _ = asset.NewTitle("Contact Lens Solution")
	item.Asset.CustomAssetTypeID = "type"
	item.TagIDs = []string{"tag-a", "tag-b"}
	item.AncestorIDs = []string{"bathroom", "cabinet"}
	filter := WorkspaceFilter{Text: "LENS", TypeID: "type", TagIDs: []string{"tag-a", "tag-b"}, LocationID: "bathroom", FromDate: "2026-09-30", ThroughDate: "2026-09-30"}
	if filter.Validate() != nil || !filter.Matches(item) {
		t.Fatal("valid combined month filter did not match")
	}
	filter.TagIDs = append(filter.TagIDs, "other")
	if filter.Matches(item) {
		t.Fatal("tag filters are not intersected")
	}
	filter.TagIDs = nil
	filter.ThroughDate = "2026-09-29"
	if filter.Validate() == nil {
		t.Fatal("reversed date range accepted")
	}
	filter.FromDate = "2026-09-01"
	if filter.Matches(item) {
		t.Fatal("month precision matched before month end")
	}
}

func TestWorkspaceComposesBrowseKindAndAvailability(t *testing.T) {
	item := workspaceItem(t, "item", "2026-09-18", true)
	item.Asset.Kind = asset.KindItem
	for _, test := range []struct {
		kind       asset.Kind
		checkout   string
		checkedOut bool
		want       bool
	}{
		{asset.KindItem, "available", false, true}, {asset.KindContainer, "available", false, false},
		{asset.KindItem, "checked_out", false, false}, {asset.KindItem, "checked_out", true, true},
	} {
		item.CheckedOut = test.checkedOut
		filter := WorkspaceFilter{Kind: test.kind, CheckoutState: ports.AssetCheckoutStateFilter(test.checkout)}
		if filter.Validate() != nil || filter.Matches(item) != test.want {
			t.Fatalf("wrong Browse composition: %+v", test)
		}
	}
}
