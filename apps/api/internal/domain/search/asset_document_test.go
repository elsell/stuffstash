package search

import "testing"

func TestMatchAssetMatchesAssignedTagNamesAndKeys(t *testing.T) {
	query, ok := NewQuery("workshop")
	if !ok {
		t.Fatal("expected valid query")
	}
	matches := MatchAsset(AssetDocument{
		Tags: []TagDocument{
			{Key: "shop-tools", DisplayName: "Workshop"},
		},
	}, query, ModeExact)
	if len(matches) != 1 || matches[0].Field != MatchFieldTagDisplayName || matches[0].Value != "Workshop" {
		t.Fatalf("expected tag display-name match, got %+v", matches)
	}

	keyQuery, ok := NewQuery("shop-tools")
	if !ok {
		t.Fatal("expected valid key query")
	}
	matches = MatchAsset(AssetDocument{
		Tags: []TagDocument{
			{Key: "shop-tools", DisplayName: "Workshop"},
		},
	}, keyQuery, ModeExact)
	if len(matches) != 1 || matches[0].Field != MatchFieldTagKey || matches[0].Value != "shop-tools" {
		t.Fatalf("expected tag key match, got %+v", matches)
	}
}
