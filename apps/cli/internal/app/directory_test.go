package app

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

func TestDirectoryCommandsRequireOnlyTheirResourceScope(t *testing.T) {
	for _, c := range []struct {
		words   []string
		scope   ports.Scope
		missing bool
	}{
		{[]string{"account", "show"}, ports.Scope{}, false},
		{[]string{"tenants", "show"}, ports.Scope{}, true},
		{[]string{"tenants", "show"}, ports.Scope{Tenant: "home"}, false},
		{[]string{"inventories", "show"}, ports.Scope{Tenant: "home"}, true},
		{[]string{"inventories", "show"}, ports.Scope{Tenant: "home", Inventory: "garage"}, false},
	} {
		o := Options{Command: c.words, Scope: c.scope}
		if err := validateCommandShape(o); err != nil {
			t.Fatal(err)
		}
		if got := missingResourceScope(o); got != c.missing {
			t.Fatalf("%v missing=%v", c.words, got)
		}
		if err := validateCommand(o); (err != nil) != c.missing {
			t.Fatalf("%v validation=%v", c.words, err)
		}
	}
}
