package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

func TestTenantListNeedsNoResourceScopeOrPicker(t *testing.T) {
	o := Options{Command: []string{"tenants", "list"}}
	if err := validateCommand(o); err != nil {
		t.Fatal(err)
	}
	if missingResourceScope(o) {
		t.Fatal("household discovery must not require a household")
	}
	r := Runner{Picker: firstScopeChoice{}}
	if _, err := r.chooseMissingScope(context.Background(), o, ports.Session{}); err != nil {
		t.Fatal(err)
	}
}
