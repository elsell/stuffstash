package app

import (
	"encoding/json"
	"slices"
	"testing"
)

// The model must see the argument families that the application actually accepts.
// A union of all fields advertises invalid operations such as assigning a new
// assetId during creation or changing a title through a move command.
func TestConversationProposalCatalogDescribesCommandArgumentFamilies(t *testing.T) {
	type schema struct {
		Properties           map[string]json.RawMessage `json:"properties"`
		Required             []string                   `json:"required"`
		AdditionalProperties bool                       `json:"additionalProperties"`
		Items                json.RawMessage            `json:"items"`
		AnyOf                []json.RawMessage          `json:"anyOf"`
		Enum                 []string                   `json:"enum"`
	}
	decode := func(raw json.RawMessage) schema {
		t.Helper()
		var result schema
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	root := decode(realtimeConversationProposalTool().Parameters)
	commands := decode(root.Properties["commands"])
	alternatives := decode(commands.Items).AnyOf
	cases := []struct {
		kind             string
		fields, required []string
	}{
		{"create_asset", []string{"title", "kind", "description", "parentAssetId", "parentCommandId"}, []string{"title"}},
		{"create_location", []string{"title", "kind", "description", "parentAssetId", "parentCommandId"}, []string{"title"}},
		{"move_asset", []string{"assetId", "parentAssetId", "parentCommandId"}, []string{"assetId"}},
		{"archive_asset", []string{"assetId"}, []string{"assetId"}},
		{"restore_asset", []string{"assetId"}, []string{"assetId"}},
		{"checkout_asset", []string{"assetId", "details"}, []string{"assetId"}},
		{"return_asset", []string{"assetId", "details"}, []string{"assetId"}},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			found := false
			for _, raw := range alternatives {
				command := decode(raw)
				discriminator := decode(command.Properties["kind"])
				if !slices.Contains(discriminator.Enum, tc.kind) {
					continue
				}
				if found {
					t.Fatal("ambiguous command schema")
				}
				found = true
				args := decode(command.Properties["arguments"])
				if args.AdditionalProperties {
					t.Fatal("undeclared arguments allowed")
				}
				if len(args.Properties) != len(tc.fields) {
					t.Fatalf("wrong argument fields: %v", args.Properties)
				}
				for _, field := range tc.fields {
					if len(args.Properties[field]) == 0 {
						t.Errorf("missing argument %s", field)
					}
				}
				for _, field := range tc.required {
					if !slices.Contains(args.Required, field) {
						t.Errorf("missing required argument %s", field)
					}
				}
			}
			if !found {
				t.Fatal("command has no discriminated argument contract")
			}
		})
	}
}
