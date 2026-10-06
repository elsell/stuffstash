package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"path/filepath"
	"testing"
)

type scopeCatalog struct{}

func (scopeCatalog) Tenants(context.Context, ports.Page) (ports.Result[[]ports.Tenant], error) {
	return ports.Result[[]ports.Tenant]{Data: []ports.Tenant{{ID: "home", Name: "Home"}}}, nil
}
func (scopeCatalog) Inventories(context.Context, ports.Scope, ports.Page) (ports.Result[[]ports.Inventory], error) {
	return ports.Result[[]ports.Inventory]{Data: []ports.Inventory{{ID: "garage", Name: "Garage"}}}, nil
}

type firstScopeChoice struct{}

func (firstScopeChoice) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	return choices[0].ID, nil
}
func TestScopePickerRemembersChoicesButNeverPromptsForScripts(t *testing.T) {
	ctx := context.Background()
	store := contextfile.Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
	session := ports.Session{Issuer: "https://identity.example", Subject: "alice", Server: "https://stash.example"}
	runner := Runner{Contexts: store, Picker: firstScopeChoice{}, ScopeAPI: func(string, string) (ports.ScopeCatalog, error) { return scopeCatalog{}, nil }}
	options := Options{Server: session.Server, Command: []string{"assets", "list"}}
	selected, err := runner.chooseMissingScope(ctx, options, session)
	if err != nil || selected.Scope.Tenant != "home" || selected.Scope.Inventory != "garage" {
		t.Fatalf("choice: %+v %v", selected.Scope, err)
	}
	current, err := (contexts.Manager{Store: store}).Current(ctx)
	if err != nil || current.Principal != contexts.Principal(session) || current.Inventory != "garage" {
		t.Fatalf("choice not remembered: %+v %v", current, err)
	}
	runner.ScopeAPI = nil
	for _, script := range []Options{{Command: options.Command, NoInput: true}, {Command: options.Command, JSON: true}} {
		got, err := runner.chooseMissingScope(ctx, script, session)
		if err != nil || got.Scope != (ports.Scope{}) {
			t.Fatalf("script tried to choose: %+v %v", got.Scope, err)
		}
	}
}
func TestScopePickerUpdatesExplicitContextNotAnotherMatchingContext(t *testing.T) {
	ctx := context.Background()
	store := contextfile.Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
	manager := contexts.Manager{Store: store}
	session := ports.Session{Server: "https://stash.example", Issuer: "https://identity.example", Subject: "alice"}
	for _, name := range []string{"first", "requested"} {
		if err := manager.Remember(ctx, contexts.Entry{Name: name, Server: session.Server, Principal: contexts.Principal(session), Tenant: "home"}); err != nil {
			t.Fatal(err)
		}
	}
	runner := Runner{Contexts: store, Picker: firstScopeChoice{}, ScopeAPI: func(string, string) (ports.ScopeCatalog, error) { return scopeCatalog{}, nil }}
	_, err := runner.chooseMissingScope(ctx, Options{Server: session.Server, Command: []string{"assets", "list"}, Scope: ports.Scope{Tenant: "home"}, Selection: contexts.Selection{Context: "requested"}}, session)
	if err != nil {
		t.Fatal(err)
	}
	current, err := manager.Current(ctx)
	if err != nil || current.Name != "requested" || current.Inventory != "garage" {
		t.Fatalf("wrong context updated: %+v %v", current, err)
	}
}
