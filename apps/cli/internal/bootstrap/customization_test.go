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
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCustomizationBoundary(t *testing.T) {
	for _, domain := range []string{"asset-types", "field-definitions"} {
		for _, scope := range []string{"household", "inventory"} {
			for _, action := range []string{"list", "show", "create", "update", "archive", "restore", "delete"} {
				t.Run(domain+"/"+scope+"/"+action, func(t *testing.T) {
					path := "/tenants/home"
					if scope == "inventory" {
						path += "/inventories/inventory"
					}
					resource := "custom-asset-types"
					body := `{"$schema":"input","key":"appliance","displayName":"Appliance","description":"","expirationEnabled":false}`
					data := `{"id":"id","tenantId":"home","scope":"` + scope + `","key":"appliance","displayName":"Appliance","description":"","expirationEnabled":false,"lifecycleState":"active"`
					if domain == "field-definitions" {
						resource = "custom-field-definitions"
						body = `{"$schema":"input","key":"grade","displayName":"Grade","type":"enum","enumOptions":["A","B"],"applicability":"custom_asset_types","customAssetTypeIds":["type"]}`
						data = `{"id":"id","tenantId":"home","scope":"` + scope + `","key":"grade","displayName":"Grade","type":"enum","enumOptions":["A","B"],"applicability":"custom_asset_types","customAssetTypeIds":["type"],"lifecycleState":"active"`
					}
					if scope == "inventory" {
						data += `,"inventoryId":"inventory"`
					}
					data += "}"
					path += "/" + resource
					method := "GET"
					if action != "list" && action != "create" {
						path += "/id"
					}
					switch action {
					case "create":
						method = "POST"
					case "update":
						method = "PATCH"
					case "archive", "restore":
						method = "PATCH"
						path += "/" + action
					case "delete":
						method = "DELETE"
					}
					if action == "update" {
						if domain == "asset-types" {
							body = `{"displayName":"Changed","description":null,"expirationEnabled":false}`
						} else {
							body = `{"displayName":"Changed","enumOptions":["A","B","C"],"applicability":"all_assets","customAssetTypeIds":["type"]}`
						}
					}
					meta := `"meta":{"requestId":"trace","tenantId":"home"}`
					if action == "list" {
						data = "[" + data + "]"
						meta = `"meta":{"requestId":"trace","tenantId":"home","pagination":{"nextCursor":"next","hasMore":true,"limit":7}}`
					}
					response := `{"$schema":"schema","data":` + data + `,` + meta + `}`
					calls, status := 0, 200
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls++
						if r.Method != method || r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
							t.Errorf("wrong route/auth %s %s", r.Method, r.URL.Path)
							w.WriteHeader(403)
							return
						}
						if action == "list" && (r.URL.Query().Get("limit") != "7" || r.URL.Query().Get("cursor") != "cursor" || r.URL.Query().Get("lifecycleState") != "all") {
							t.Error("lost filters")
						}
						if action == "create" || action == "update" {
							got, _ := io.ReadAll(r.Body)
							if string(got) != body {
								t.Errorf("body changed: %s", got)
							}
						}
						if status != 200 {
							w.WriteHeader(status)
							io.WriteString(w, "private-denial")
							return
						}
						if action == "delete" {
							w.WriteHeader(204)
							return
						}
						io.WriteString(w, response)
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
							return filepath.Join(dir, "contexts.json")
						case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
							return "true"
						}
						return ""
					}
					args := []string{domain, action}
					if action != "list" && action != "create" {
						args = append(args, "id")
					}
					args = append(args, "--scope", scope, "--tenant", "home", "--json", "--no-input")
					if scope == "inventory" {
						args = append(args, "--inventory", "inventory")
					}
					if action == "list" {
						args = append(args, "--limit", "7", "--cursor", "cursor", "--lifecycle", "all")
					}
					if action == "create" || action == "update" {
						file := filepath.Join(dir, "input.json")
						os.WriteFile(file, []byte(body), 0600)
						args = append(args, "--input", file)
					}
					var out, diag bytes.Buffer
					if action != "list" && action != "show" {
						if code := Run(context.Background(), args, getenv, &out, &diag); code != 2 || calls != 0 {
							t.Fatalf("unconfirmed change: %d %s", code, &diag)
						}
						args = append(args, "--yes")
					}
					out.Reset()
					diag.Reset()
					if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != 1 {
						t.Fatalf("command: %d %s", code, &diag)
					}
					if action != "delete" {
						var got, want any
						json.Unmarshal(out.Bytes(), &got)
						json.Unmarshal([]byte(response), &want)
						if action == "list" {
							want.(map[string]any)["pagination"] = want.(map[string]any)["meta"].(map[string]any)["pagination"]
						}
						if !reflect.DeepEqual(got, want) {
							t.Fatalf("result mismatch: %s want %s", &out, response)
						}
					}
					for _, denial := range []int{401, 403, 409, 500} {
						status = denial
						before := calls
						out.Reset()
						diag.Reset()
						if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "private-denial") {
							t.Fatalf("denial/retry: %d %s", code, &diag)
						}
					}
				})
			}
		}
	}
}
