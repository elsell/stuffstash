package bootstrap

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"os"
	"strings"
	"testing"
)

func TestHelpCatalogCoversImplementedAPICommands(t *testing.T) {
	raw, err := os.ReadFile("../../api-coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Operations map[string]struct {
			Commands []string `json:"commands"`
		} `json:"operations"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for id, operation := range manifest.Operations {
		for _, path := range operation.Commands {
			entry, ok := helpCommand(strings.Fields(path))
			if !ok || entry.Path != path {
				t.Errorf("implemented %s has no help: %s", id, path)
			}
		}
	}
	flags := app.HelpOptions()
	seen := map[string]bool{}
	for _, entry := range helpCatalog() {
		if entry.Path == "" || entry.Summary == "" || scopeHelp(entry.Scope) == "" || seen[entry.Path] {
			t.Errorf("incomplete or duplicate help: %+v", entry)
		}
		seen[entry.Path] = true
		names := map[string]bool{}
		for _, name := range commandHelpOptions(entry) {
			if _, ok := flags[name]; !ok || names[name] {
				t.Errorf("invalid help option %s for %s", name, entry.Path)
			}
			names[name] = true
		}
	}
}

func TestAllPagesHelpOptionsAreExecutable(t *testing.T) {
	for _, entry := range helpCatalog() {
		command := strings.Fields(entry.Path + " " + entry.Arguments)
		if !app.SupportsAllPages(command) {
			continue
		}
		args := append(command, "--all", "--no-input", "--json")
		if entry.Scope == helpConnector {
			args = append(args, "--connector", "connector")
		}
		parsed, err := app.Parse(args, func(string) string { return "" })
		if err != nil || !parsed.AllPages {
			t.Errorf("advertised traversal rejected for %s: %v", entry.Path, err)
		}
		found := false
		for _, name := range commandHelpOptions(entry) {
			if name == "all" {
				found = true
			}
		}
		if !found || !strings.Contains(commandHelpText(entry), "not an atomic snapshot") {
			t.Errorf("missing traversal help for %s", entry.Path)
		}
	}
}
