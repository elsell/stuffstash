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

func TestFuzzySearchCombinesTermsAcrossFieldsOnTheSameAsset(t *testing.T) {
	document := AssetDocument{Title: "3–6 months clothes", Tags: []TagDocument{{Key: "baby", DisplayName: "Baby"}}}
	for _, query := range []string{"baby clothes", "CLOTHES   baby", "baby baby clothes"} {
		matches := MatchAsset(document, Query(query), ModeFuzzy)
		if len(matches) != 3 {
			t.Fatalf("cross-field evidence lost or duplicated for %q: %+v", query, matches)
		}
	}
	if matches := MatchAsset(document, Query("baby stroller"), ModeFuzzy); len(matches) != 0 {
		t.Fatalf("partial term match passed: %+v", matches)
	}
	if matches := MatchAsset(document, Query("baby clothes"), ModeExact); len(matches) != 0 {
		t.Fatalf("exact equality became term matching: %+v", matches)
	}
	if matches := MatchAsset(AssetDocument{Title: "Baby clothes"}, Query("baby clothes"), ModeExact); len(matches) != 1 {
		t.Fatal("exact phrase lookup regressed")
	}
}
