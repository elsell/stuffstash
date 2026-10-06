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

func TestConnectorInspectionFullReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := "/tenants/home/inventories/garage/print-connectors"
		if r.Header.Get("Authorization") != "Bearer owner" || !strings.HasPrefix(r.URL.Path, prefix) {
			w.WriteHeader(403)
			return
		}
		if r.Method != "GET" {
			t.Error("wrong method")
		}
		data := `{"id":"connector","name":"Kitchen","generation":9007199254740993,"authorizationPending":false,"availability":"online","state":"active","lastSeenAt":"2026-10-05T10:00:00Z","reportReceivedAt":"2026-10-05T10:00:00Z","printerIds":["printer"],"report":{"architecture":"amd64","platform":"linux","commit":"abc","version":"1","adapters":[{"id":"brother","completionEvidence":"ack","contractVersions":[1],"formats":["png"],"wake":true,"media":[{"id":"preset","version":2}]}]}}`
		if r.URL.Path == prefix {
			if r.URL.Query().Get("cursor") != "before" || r.URL.Query().Get("limit") != "1" {
				t.Error("lost paging")
			}
			data = "[" + data + "]"
		}
		io.WriteString(w, `{"$schema":"connector-schema","data":`+data+`,"meta":{"requestId":"trace","pagination":{"hasMore":true,"nextCursor":"after"}}}`)
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

	for _, action := range [][]string{{"show", "connector"}, {"list", "--limit", "1", "--cursor", "before"}} {
		command := append([]string{"connectors", "print"}, action...)
		command = append(command, "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("connector: %d %s", code, &diagnostic)
		}
		for _, field := range []string{`"generation":9007199254740993`, `"authorizationPending":false`, `"printerIds":["printer"]`, `"architecture":"amd64"`, `"platform":"linux"`, `"commit":"abc"`, `"completionEvidence":"ack"`, `"contractVersions":[1]`, `"formats":["png"]`, `"wake":true`, `"media":[{"id":"preset","version":2}]`, `"reportReceivedAt":"2026-10-05T10:00:00Z"`, `"requestId":"trace"`, `"$schema":"connector-schema"`} {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
			if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
				t.Fatal("scope bypass")
			}
		}
	}
}
