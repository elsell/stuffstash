package scopeselection

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

type catalog struct{ repeat bool }

func (c catalog) Tenants(_ context.Context, p ports.Page) (ports.Result[[]ports.Tenant], error) {
	next := "next"
	if p.Cursor == "" || c.repeat {
		return ports.Result[[]ports.Tenant]{Data: []ports.Tenant{{ID: "other", Name: "Other"}}, Pagination: &ports.Pagination{HasMore: true, NextCursor: &next}}, nil
	}
	return ports.Result[[]ports.Tenant]{Data: []ports.Tenant{{ID: "home", Name: "Home"}}}, nil
}
func (c catalog) Inventories(_ context.Context, s ports.Scope, _ ports.Page) (ports.Result[[]ports.Inventory], error) {
	if s.Tenant != "home" {
		return ports.Result[[]ports.Inventory]{}, ports.Failure("forbidden", "No access.")
	}
	return ports.Result[[]ports.Inventory]{Data: []ports.Inventory{{ID: "garage", Name: "Garage"}}}, nil
}

type chooser struct{ choices []string }

func (c *chooser) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	id := c.choices[0]
	c.choices = c.choices[1:]
	return id, nil
}
func TestChooseAuthorizedScopeAcrossPages(t *testing.T) {
	picker := &chooser{choices: []string{"home", "garage"}}
	got, err := Choose(context.Background(), catalog{}, picker, ports.Scope{}, true)
	if err != nil || got.Tenant != "home" || got.Inventory != "garage" {
		t.Fatalf("scope: %+v %v", got, err)
	}
}
func TestScopeSelectionRejectsUnknownChoiceAndRepeatedCursor(t *testing.T) {
	for _, test := range []struct {
		catalog catalog
		choice  string
	}{{catalog{}, "unauthorized"}, {catalog{repeat: true}, "home"}} {
		if _, err := Choose(context.Background(), test.catalog, &chooser{choices: []string{test.choice}}, ports.Scope{}, true); err == nil {
			t.Fatal("invalid catalog/choice accepted")
		}
	}
}
