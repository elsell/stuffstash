package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
	"testing"
)

func TestAssetCreateRetryKeysMatchServerSupport(t *testing.T) {
	for _, body := range []string{`{"title":"Box","kind":"container"}`, `{"title":"Box","kind":"container","printLabel":null}`} {
		r := Runner{InputFiles: requestInput(body)}
		_, err := r.prepareInput(context.Background(), Options{Command: []string{"assets", "create"}, InputPath: "-", IdempotencyKey: "retry"})
		if err == nil {
			t.Fatal("ordinary create accepted retry key")
		}
	}
	r := Runner{InputFiles: requestInput(`{"title":"Box","kind":"container","printLabel":{"printerId":"printer"}}`)}
	o, err := r.prepareInput(context.Background(), Options{Command: []string{"assets", "create"}, InputPath: "-", IdempotencyKey: "retry"})
	if err != nil || !assetPrintRequested(o) {
		t.Fatalf("print retry rejected: %v", err)
	}
	for _, print := range []bool{false, true} {
		err := assetCreateFailure(Options{PrintLabel: print}, ports.Failure("network", "connection failed"))
		want := "Run assets list"
		if print {
			want = "same request key"
		}
		if !strings.Contains(err.Error(), want) {
			t.Fatal(err)
		}
	}
	denied := ports.Failure("forbidden", "Access denied.")
	if assetCreateFailure(Options{}, denied) != denied {
		t.Fatal("authorization failure replaced")
	}
}
