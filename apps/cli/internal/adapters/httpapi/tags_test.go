package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTagsSDKRequestsKeepScopeFieldsAndDenials(t *testing.T) {
	tag := `{"id":"tag","tenantId":"home","inventoryId":"garage","key":"tools","displayName":"Tools","color":"","lifecycleState":"active","createdAt":"created","updatedAt":"updated"}`
	for _, action := range []string{"list", "create", "update", "delete"} {
		t.Run(action, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					io.WriteString(w, `{"error":{"message":"private detail"}}`)
					return
				}
				path := "/tenants/home/inventories/garage/tags"
				method := "GET"
				if action == "create" {
					method = "POST"
				}
				if action == "update" {
					method = "PATCH"
					path += "/tag"
				}
				if action == "delete" {
					method = "DELETE"
					path += "/tag"
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("wrong request: %s %s", r.Method, r.URL.Path)
				}
				if action == "create" || action == "update" {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"color":""}` {
						t.Errorf("body changed: %s", b)
					}
				}
				data := tag
				if action == "list" {
					data = "[" + tag + "]"
					if r.URL.Query().Get("limit") != "0" || r.URL.Query().Get("cursor") != "next" {
						t.Errorf("wrong query: %s", r.URL)
					}
				}
				io.WriteString(w, `{"data":`+data+`,"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			for _, token := range []string{"owner", "viewer", ""} {
				api, _ := New(server.URL, token, server.Client())
				var result any
				var err error
				if action == "list" {
					result, err = api.Tags(context.Background(), ports.Scope{Tenant: "home", Inventory: "garage"}, ports.Page{Cursor: "next"})
				} else {
					result, err = api.ChangeTag(context.Background(), ports.Scope{Tenant: "home", Inventory: "garage"}, ports.TagAction(action), "tag", []byte(`{"color":""}`))
				}
				if token != "owner" {
					if err == nil || strings.Contains(err.Error(), "private") {
						t.Fatalf("unsafe denial %v", err)
					}
					continue
				}
				data, _ := json.Marshal(result)
				if err != nil || !strings.Contains(string(data), `"color":""`) || !strings.Contains(string(data), `"createdAt":"created"`) || !strings.Contains(string(data), `"requestId":"trace"`) {
					t.Fatalf("lost fields %s %v", data, err)
				}
			}
		})
	}
}
