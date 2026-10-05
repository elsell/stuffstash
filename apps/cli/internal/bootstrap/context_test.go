package bootstrap

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

func TestLocalContextListWithoutServerOrLogin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "contexts.json")
	getenv := func(key string) string {
		if key == "STUFF_STASH_CLI_CONFIG_FILE" {
			return path
		}
		return ""
	}
	var out, diagnostics bytes.Buffer
	code := Run(context.Background(), []string{"context", "list", "--json", "--no-input"}, getenv, &out, &diagnostics)
	if code != 0 || out.String() != "[]\n" || diagnostics.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
}

func TestCanceledContextCommandExits130(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, diagnostics bytes.Buffer
	getenv := func(key string) string {
		if key == "STUFF_STASH_CLI_CONFIG_FILE" {
			return filepath.Join(t.TempDir(), "contexts.json")
		}
		return ""
	}
	if code := Run(ctx, []string{"context", "list", "--json"}, getenv, &out, &diagnostics); code != 130 {
		t.Fatalf("cancellation exit=%d: %s", code, diagnostics.String())
	}
}
