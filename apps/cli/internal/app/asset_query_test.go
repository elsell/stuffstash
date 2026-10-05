package app

import "testing"

func TestAssetListFiltersValidateWithoutChangingOtherCommands(t *testing.T) {
	o, err := Parse([]string{"assets", "list", "--lifecycle", "all", "--sort", "updated_desc"}, func(string) string { return "" })
	if err != nil || o.Lifecycle != "all" || o.Sort != "updated_desc" {
		t.Fatalf("filters: %+v %v", o, err)
	}
	for _, args := range [][]string{{"assets", "list", "--lifecycle", "unknown"}, {"assets", "list", "--sort", "title"}, {"assets", "show", "id", "--lifecycle", "all"}} {
		if _, err := Parse(args, func(string) string { return "" }); err == nil {
			t.Fatalf("unsupported filter accepted: %v", args)
		}
	}
}
