package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"path/filepath"
	"testing"
)

func TestVoiceProviderPickerNeedsOnlyHousehold(t *testing.T) {
	runner := Runner{Contexts: contextfile.Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}, Picker: firstScopeChoice{}, ScopeAPI: func(string, string) (ports.ScopeCatalog, error) { return scopeCatalog{}, nil }}
	selected, err := runner.chooseMissingScope(context.Background(), Options{Server: "https://stash.example", Command: []string{"voice-provider", "show"}}, ports.Session{Server: "https://stash.example", Issuer: "https://identity.example", Subject: "owner"})
	if err != nil || selected.Scope.Tenant != "home" || selected.Scope.Inventory != "" {
		t.Fatalf("household picker: %+v %v", selected.Scope, err)
	}
}
func TestInvalidVoiceProviderShapeBeforeCredentials(t *testing.T) {
	for _, command := range [][]string{{"voice-provider"}, {"voice-provider", "set"}, {"voice-provider", "show", "extra"}} {
		if err := (Runner{}).Run(context.Background(), Options{Command: command}); err == nil {
			t.Fatalf("accepted %v", command)
		}
	}
}
