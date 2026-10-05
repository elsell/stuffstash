package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestAccessGrantRemovalConfirmationAndScope(t *testing.T) {
	for _, action := range []string{"viewer", "editor"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			creating := false
			reading := false
			listing := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				expected := "/tenants/home/inventories/garage/access-grants/member/" + action
				if listing || creating {
					expected = "/tenants/home/inventories/garage/access-grants"
				}
				if r.URL.Path != expected || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if creating {
					if r.Method != "POST" {
						t.Error("wrong create method")
					}
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["principalId"] != "member" || body["relationship"] != action {
						t.Error("wrong grant body")
					}
					io.WriteString(w, `{"data":{"tenantId":"home","inventoryId":"garage","principalId":"member","relationship":"`+action+`"},"meta":{}}`)
					return
				}

				if reading {
					if r.Method != "GET" {
						t.Error("wrong read method")
					}
					data := `{"tenantId":"home","inventoryId":"garage","principalId":"member","relationship":"` + action + `"}`
					if listing {
						if r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "next" {
							t.Error("pagination lost")
						}
						data = "[" + data + "]"
					}
					io.WriteString(w, `{"data":`+data+`,"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
					return
				}
				if r.Method != "DELETE" {
					t.Error("wrong method")
				}
				w.WriteHeader(204)
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

			scope := []string{"--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
			command := append([]string{"access-grants", "remove", "member", action}, scope...)
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != 0 {
				t.Fatal("unconfirmed operation reached API")
			}
			command = append(command, "--yes")
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 {
				t.Fatalf("operation: %d %d %s %s", code, calls, &out, &diagnostic)
			}
			for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
				if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("scope boundary bypassed")
				}
			}
			creating = true
			input := filepath.Join(dir, "grant.json")
			os.WriteFile(input, []byte(`{"principalId":"member","relationship":"`+action+`"}`), 0600)
			create := append([]string{"access-grants", "create", "--input", input}, scope...)
			beforeCreate := calls
			if code := Run(context.Background(), create, getenv, &out, &diagnostic); code != 2 || calls != beforeCreate {
				t.Fatal("unconfirmed grant created")
			}
			create = append(create, "--yes")
			if code := Run(context.Background(), create, getenv, &out, &diagnostic); code != 0 || calls != beforeCreate+1 {
				t.Fatalf("create: %d %s", code, &diagnostic)
			}
			for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
				if code := Run(context.Background(), append(create, override...), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("create scope bypass")
				}
			}
			os.WriteFile(input, []byte(`{"principalId":"member","relationship":"owner"}`), 0600)
			beforeCreate = calls
			if code := Run(context.Background(), create, getenv, &out, &diagnostic); code != 2 || calls != beforeCreate {
				t.Fatal("invalid create reached API")
			}
			creating = false

			reading = true
			for _, mode := range []string{"show", "list"} {
				listing = mode == "list"
				readCommand := append([]string{"access-grants", "show", "member", action}, scope...)
				if listing {
					readCommand = append([]string{"access-grants", "list", "--limit", "2", "--cursor", "next"}, scope...)
				}
				out.Reset()
				diagnostic.Reset()
				if code := Run(context.Background(), readCommand, getenv, &out, &diagnostic); code != 0 {
					t.Fatalf("read: %d %s", code, &diagnostic)
				}
				for _, field := range []string{`"principalId":"member"`, `"relationship":"` + action + `"`, `"requestId":"trace"`} {
					if !strings.Contains(out.String(), field) {
						t.Fatalf("lost grant: %s", &out)
					}
				}
				if listing && !strings.Contains(out.String(), `"nextCursor":"more"`) {
					t.Fatal("lost cursor")
				}
				for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
					if code := Run(context.Background(), append(readCommand, override...), getenv, &out, &diagnostic); code == 0 {
						t.Fatal("read scope boundary bypassed")
					}
				}
			}
			before := calls
			if code := Run(context.Background(), append([]string{"access-grants", "remove", "member", "owner", "--yes"}, scope...), getenv, &out, &diagnostic); code != 2 || calls != before {
				t.Fatal("invalid role reached API")
			}

		})
	}
}
