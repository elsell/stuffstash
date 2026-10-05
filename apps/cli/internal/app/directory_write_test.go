package app

import (
	"context"
	"testing"
)

type requestInput []byte

func (b requestInput) Read(context.Context, string) ([]byte, error) { return b, nil }
func TestDirectoryBodyValidationBeforeWrite(t *testing.T) {
	for _, body := range []string{`[]`, `null`, `{"name":"x"} {}`, `{`} {
		r := Runner{InputFiles: requestInput(body)}
		if _, err := r.prepareInput(context.Background(), Options{Command: []string{"tenants", "create"}, InputPath: "-"}); err == nil {
			t.Fatalf("invalid body accepted: %s", body)
		}
	}
	body := `{"name":null}`
	r := Runner{InputFiles: requestInput(body)}
	got, err := r.prepareInput(context.Background(), Options{Command: []string{"tenants", "update"}, InputPath: "-"})
	if err != nil || string(got.RequestBody) != body {
		t.Fatalf("body changed: %s %v", got.RequestBody, err)
	}
	if _, err := r.prepareInput(context.Background(), Options{Command: []string{"assets", "list"}, InputPath: "-"}); err == nil {
		t.Fatal("ignored body accepted")
	}
}

func TestDirectoryWritesDoNotPromiseUnsupportedIdempotency(t *testing.T) {
	r := Runner{}
	if _, err := r.prepareInput(context.Background(), Options{Command: []string{"tenants", "create"}, ConnectorName: "Home", IdempotencyKey: "retry"}); err == nil {
		t.Fatal("unsupported retry key accepted")
	}
}
