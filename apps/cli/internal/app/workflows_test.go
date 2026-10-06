package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"path/filepath"
	"testing"
)

func TestWorkflowPickerNeedsOnlyHousehold(t *testing.T) {
	runner := Runner{Contexts: contextfile.Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}, Picker: firstScopeChoice{}, ScopeAPI: func(string, string) (ports.ScopeCatalog, error) { return scopeCatalog{}, nil }}
	for _, command := range [][]string{{"workflows", "create"}, {"workflows", "revisions", "create", "wf"}, {"workflows", "activate", "wf"}, {"workflows", "list"}, {"workflows", "show", "id"}, {"workflows", "revisions", "list", "id"}, {"workflows", "revisions", "show", "id", "rev"}, {"workflows", "selection", "show"}} {
		selected, err := runner.chooseMissingScope(context.Background(), Options{Server: "https://stash.example", Command: command}, ports.Session{Server: "https://stash.example", Issuer: "https://id.example", Subject: "owner"})
		if err != nil || selected.Scope.Tenant != "home" || selected.Scope.Inventory != "" {
			t.Fatalf("household picker for %v: %+v %v", command, selected.Scope, err)
		}
	}
}

func TestInvalidWorkflowShapeBeforeCredentials(t *testing.T) {
	for _, command := range [][]string{{"workflows"}, {"workflows", "show"}, {"workflows", "revisions", "show", "id"}, {"workflows", "revisions", "list", ""}, {"workflows", "selection", "show", "extra"}} {
		if err := (Runner{}).Run(context.Background(), Options{Command: command}); err == nil {
			t.Fatalf("accepted %v", command)
		}
	}
}
