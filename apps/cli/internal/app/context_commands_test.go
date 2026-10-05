package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"path/filepath"
	"testing"
)

type contextOutput struct{ value any }

func (o *contextOutput) Result(value any) error { o.value = value; return nil }
func (*contextOutput) Notice(string) error      { return nil }
func (*contextOutput) Error(string, string)     {}
func TestContextCommandsNeedNoServerOrCredentials(t *testing.T) {
	ctx := context.Background()
	store := contextfile.Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
	manager := contexts.Manager{Store: store}
	if err := manager.Remember(ctx, contexts.Entry{Name: "home", Server: "https://home.example", Principal: "alice", Tenant: "home", Inventory: "garage"}); err != nil {
		t.Fatal(err)
	}
	output := &contextOutput{}
	runner := Runner{Contexts: store, Output: output}
	for _, command := range [][]string{{"context", "list"}, {"context", "current"}, {"context", "use", "home"}, {"context", "delete", "home"}} {
		if err := runner.Run(ctx, Options{Command: command}); err != nil {
			t.Fatalf("%v: %v", command, err)
		}
	}
	for _, command := range [][]string{{"context", "current"}, {"context", "use", "missing"}, {"context", "delete"}, {"context", "list", "extra"}} {
		if err := runner.Run(ctx, Options{Command: command}); err == nil {
			t.Fatalf("invalid context command succeeded: %v", command)
		}
	}
}

func TestInvalidCommandDoesNotLoadCredentials(t *testing.T) {
	// Missing credential/auth ports make any premature access fail immediately.
	runner := Runner{Contexts: contextfile.Store{Path: filepath.Join(t.TempDir(), "contexts.json")}}
	for _, command := range [][]string{{"unknown", "action"}, {"assets", "create"}, {"assets", "move"}} {
		if err := runner.Run(context.Background(), Options{Command: command}); err == nil {
			t.Fatalf("invalid command accepted: %v", command)
		}
	}
}
