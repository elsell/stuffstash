package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"path/filepath"
	"testing"
)

func TestEvaluationHouseholdPickerAndValidation(t *testing.T) {
	ctx := context.Background()
	session := ports.Session{Issuer: "https://id.example", Subject: "owner", Server: "https://stash.example"}
	for _, command := range [][]string{{"evaluation", "cases", "list"}, {"evaluation", "cases", "show", "case"}, {"evaluation", "revisions", "list", "case"}, {"evaluation", "revisions", "show", "case", "revision"}, {"evaluation", "runs", "list"}, {"evaluation", "runs", "show", "run"}, {"evaluation", "cases", "create"}, {"evaluation", "revisions", "create", "case"}, {"evaluation", "runs", "create"}} {
		store := contextfile.Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
		runner := Runner{Contexts: store, Picker: firstScopeChoice{}, ScopeAPI: func(string, string) (ports.ScopeCatalog, error) { return scopeCatalog{}, nil }}
		o := Options{Server: session.Server, Command: command}
		selected, err := runner.chooseMissingScope(ctx, o, session)
		if err != nil || selected.Scope.Tenant != "home" || selected.Scope.Inventory != "" {
			t.Fatalf("household picker: %+v %v", selected.Scope, err)
		}
		if err := validateEvaluation(selected, true); err != nil {
			t.Fatal(err)
		}
		for _, script := range []Options{{Command: command, JSON: true}, {Command: command, NoInput: true}} {
			selected, err = runner.chooseMissingScope(ctx, script, session)
			if err != nil || selected.Scope.Tenant != "" || validateEvaluation(selected, true) == nil {
				t.Fatal("script selected scope")
			}
		}
	}
	for _, args := range [][]string{{"evaluation"}, {"evaluation", "cases", "show"}, {"evaluation", "revisions", "show", "case"}, {"evaluation", "runs", "create", "extra"}, {"evaluation", "cases", "list", "extra"}} {
		if err := validateEvaluation(Options{Command: args}, false); err == nil {
			t.Fatalf("invalid grammar: %v", args)
		}
	}
}
