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

func TestImportJobReadsPreserveEvidenceAndScope(t *testing.T) {
	calls := 0
	listing := false
	deleting := false
	path := "/tenants/home/inventories/garage/imports/jobs/job"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		expected := path
		if listing {
			expected = strings.TrimSuffix(path, "/job")
		}
		if r.URL.Path != expected || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if deleting {
			if r.Method != "DELETE" {
				t.Error("wrong delete method")
			}
			w.WriteHeader(204)
			return
		}
		if r.Method != "GET" {
			t.Error("wrong method")
		}
		data := `{"id":"job","status":"completed","createdAt":"today","updatedAt":"today","actorId":"owner","actor":{"id":"owner","email":"owner@example.test"},"counts":{"assetsCreated":9007199254740993},"source":{"type":"json","name":"Backup","imageImport":"all","allowInsecureTLS":false,"allowPrivateNetwork":false},"progress":{"phase":"done","done":2,"total":2},"progressHistory":[{"phase":"assets","done":1,"total":2}],"preview":{"assets":[{"title":"Drill","kind":"item","archived":false}],"assetsTruncated":true},"resources":[{"resourceId":"asset","sourceEntityId":"source","resourceType":"asset","sourceEntityType":"asset","createdAt":"today"}],"messages":[{"code":"warning","severity":"warning","summary":"Review source"}]}`
		if listing {
			data = `{"jobs":[` + data + `]}`
		}
		io.WriteString(w, `{"data":`+data+`,"meta":{"requestId":"trace"}}`)

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

	command := []string{"import-jobs", "show", "job", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
		t.Fatalf("read: %d %s", code, &diagnostic)
	}
	for _, field := range []string{`"assetsCreated":9007199254740993`, `"email":"owner@example.test"`, `"assetsTruncated":true`, `"resourceId":"asset"`, `"summary":"Review source"`, `"phase":"assets"`, `"requestId":"trace"`} {
		if !strings.Contains(out.String(), field) {
			t.Fatalf("lost job field %s", &out)
		}
	}
	for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
		if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
			t.Fatal("scope bypass")
		}
	}
	before := calls
	if code := Run(context.Background(), append(command, "--cursor", "next"), getenv, &out, &diagnostic); code != 2 || calls != before {
		t.Fatal("unsupported cursor reached API")
	}
	listing = true
	listCommand := []string{"import-jobs", "list", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), listCommand, getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), `"jobs":[`) || !strings.Contains(out.String(), `"assetsCreated":9007199254740993`) {
		t.Fatalf("list: %d %s %s", code, &out, &diagnostic)
	}
	listing = false
	deleting = true
	remove := append([]string(nil), command...)
	remove[1] = "delete"
	before = calls
	if code := Run(context.Background(), remove, getenv, &out, &diagnostic); code != 2 || calls != before {
		t.Fatal("unconfirmed history removal")
	}
	remove = append(remove, "--yes")
	if code := Run(context.Background(), remove, getenv, &out, &diagnostic); code != 0 || calls != before+1 {
		t.Fatalf("remove: %d %s", code, &diagnostic)
	}
	for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
		if code := Run(context.Background(), append(remove, override...), getenv, &out, &diagnostic); code == 0 {
			t.Fatal("remove scope bypass")
		}
	}

}
