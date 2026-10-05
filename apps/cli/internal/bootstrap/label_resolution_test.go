package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLabelResolutionCompleteIdentity(t *testing.T) {
	const instance = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const label = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	calls := 0
	mismatch := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(401)
			return
		}
		if r.Method != "GET" {
			t.Error("wrong method")
		}
		if r.URL.Path == "/instance" {
			io.WriteString(w, `{"data":{"instanceId":"`+instance+`","protocolVersion":1}}`)
			return
		}
		calls++
		if r.URL.Path != "/labels/v1/"+instance+"/"+label {
			t.Error("wrong label route")
		}
		id := label
		if mismatch {
			id = instance
		}
		io.WriteString(w, `{"$schema":"label-schema","data":{"instanceId":"`+instance+`","labelId":"`+id+`","url":"https://canonical.example/l/v1/`+instance+"/"+label+`","tenantId":"home","inventoryId":"garage","assetId":"asset","lifecycleState":"active"},"meta":{"requestId":"trace"}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	if err := store.Save(context.Background(), ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	getenv := func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "config", "contexts.json")
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}

	command := []string{"labels", "resolve", "https://untrusted.invalid/l/v1/" + instance + "/" + label, "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 {
		t.Fatalf("resolve: %d %s", code, &diagnostic)
	}
	for _, field := range []string{`"instanceId":"` + instance + `"`, `"labelId":"` + label + `"`, `"url":"https://canonical.example/`, `"tenantId":"home"`, `"inventoryId":"garage"`, `"assetId":"asset"`, `"lifecycleState":"active"`, `"requestId":"trace"`, `"$schema":"label-schema"`} {
		if !strings.Contains(out.String(), field) {
			t.Fatalf("missing %s: %s", field, &out)
		}
	}
	mismatch = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || out.Len() != 0 {
		t.Fatal("mismatched identity accepted")
	}
}
