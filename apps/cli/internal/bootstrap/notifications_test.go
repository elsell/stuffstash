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

func TestNotificationCommandsBoundary(t *testing.T) {
	for _, action := range []string{"list", "show", "unread-count", "read", "unread", "read-all"} {
		t.Run(action, func(t *testing.T) {
			path := "/tenants/home/inventories/garage/notifications"
			method := "GET"
			switch action {
			case "show":
				path += "/note"
			case "unread-count", "read-all":
				path += "/" + action
			case "read", "unread":
				path += "/note/read"
			}
			if action == "read" || action == "read-all" {
				method = "PUT"
			}
			if action == "unread" {
				method = "DELETE"
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != method {
					t.Error("wrong method")
				}
				if action == "list" && (r.URL.Query().Get("unreadOnly") != "true" || r.URL.Query().Get("limit") != "2") {
					t.Error("list filters lost")
				}
				if (action == "list" || action == "read-all" || action == "unread-count") && r.URL.Query().Get("cursor") != "next" {
					t.Error("cursor lost")
				}
				data := `{"id":"note","assetId":"asset","title":"Milk","parentAssetId":"fridge","customAssetTypeId":"food","expirationDate":"2026-10-05","expirationPrecision":"day","milestone":"expired","createdAt":"2026-10-05T00:00:00Z","parentTrail":[{"assetId":"fridge","title":"Fridge","kind":"container"}],"parentTrailIncomplete":true}`
				switch action {
				case "list":
					data = "[" + data + "]"
				case "unread-count":
					data = `{"count":9}`
				case "read-all":
					data = `{"complete":false}`
				case "read":
					data = `{"id":"note","read":true}`
				case "unread":
					data = `{"id":"note","read":false}`
				}
				io.WriteString(w, `{"data":`+data+`,"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
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

			command := []string{"notifications", action}
			if action == "show" || action == "read" || action == "unread" {
				command = append(command, "note")
			}
			if action == "list" {
				command = append(command, "--unread-only", "--limit", "2")
			}
			if action == "list" || action == "unread-count" || action == "read-all" {
				command = append(command, "--cursor", "next")
			}
			command = append(command, "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"requestId":"trace"`) {
				t.Fatalf("command failed: %d %d %s %s", code, calls, &out, &diagnostic)
			}
			before := calls
			var invalidOutput, invalidError bytes.Buffer
			if code := Run(context.Background(), append(command, "--print-label"), getenv, &invalidOutput, &invalidError); code != 2 || calls != before {
				t.Fatal("unsupported print action reached API")
			}
			expected := map[string]string{"list": `"parentTrailIncomplete":true`, "show": `"parentTrailIncomplete":true`, "unread-count": `"count":9`, "read-all": `"complete":false`, "read": `"read":true`, "unread": `"read":false`}[action]
			if !strings.Contains(out.String(), expected) {
				t.Fatalf("response lost fields: %s", &out)
			}
			if action == "read-all" {
				if !strings.Contains(out.String(), `"nextCursor":"more"`) {
					t.Fatal("continuation lost")
				}
				human := make([]string, 0, len(command))
				for _, arg := range command {
					if arg != "--json" {
						human = append(human, arg)
					}
				}
				var humanOut, humanErr bytes.Buffer
				if code := Run(context.Background(), human, getenv, &humanOut, &humanErr); code != 0 || !strings.Contains(humanOut.String(), `"false"`) || !strings.Contains(humanOut.String(), `--cursor "more"`) {
					t.Fatalf("incomplete read-all hidden: %s %s", &humanOut, &humanErr)
				}
			}
			for _, scope := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
				out.Reset()
				diagnostic.Reset()
				if code := Run(context.Background(), append(command, scope...), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("wrong scope succeeded")
				}
			}
		})
	}
}
