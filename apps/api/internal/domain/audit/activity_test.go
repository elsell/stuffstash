package audit

import "testing"

func TestAssetActivityCategorySeparatesChangesFromReads(t *testing.T) {
	tests := []struct {
		action Action
		want   AssetActivityCategory
	}{
		{ActionAssetCreated, AssetActivityCategoryChange},
		{ActionAssetUpdated, AssetActivityCategoryChange},
		{ActionAssetMoved, AssetActivityCategoryChange},
		{ActionAssetArchived, AssetActivityCategoryChange},
		{ActionAssetRestored, AssetActivityCategoryChange},
		{ActionAssetCheckedOut, AssetActivityCategoryChange},
		{ActionAssetReturned, AssetActivityCategoryChange},
		{ActionAssetReturnDetailsUpdated, AssetActivityCategoryChange},
		{ActionAssetViewed, AssetActivityCategoryRead},
		{ActionAttachmentListed, AssetActivityCategoryRead},
		{ActionAttachmentContentDownloaded, AssetActivityCategoryRead},
	}
	for _, tt := range tests {
		if got := tt.action.AssetActivityCategory(); got != tt.want {
			t.Errorf("%s category = %s, want %s", tt.action, got, tt.want)
		}
	}
}

func TestAssetActivityChangeActionsAreStableDomainValues(t *testing.T) {
	actions := AssetActivityChangeActions()
	if len(actions) == 0 {
		t.Fatal("expected change actions")
	}
	for _, action := range actions {
		if action.AssetActivityCategory() != AssetActivityCategoryChange {
			t.Fatalf("change action %s has category %s", action, action.AssetActivityCategory())
		}
	}
}
