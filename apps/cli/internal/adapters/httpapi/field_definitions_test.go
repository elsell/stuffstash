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

func TestFieldDefinitionTransportRoutesAndScope(t *testing.T) {
	for _, level := range []ports.DefinitionLevel{ports.HouseholdDefinition, ports.InventoryDefinition} {
		for _, action := range []string{"list", "show", "create", "update", "archive", "restore", "delete"} {
			t.Run(string(level)+"/"+action, func(t *testing.T) {
				path := "/tenants/home"
				if level == ports.InventoryDefinition {
					path += "/inventories/tools"
				}
				path += "/custom-field-definitions"
				method := "GET"
				if action != "list" && action != "create" {
					path += "/type"
				}
				if action == "archive" || action == "restore" {
					path += "/" + action
				}
				if action == "create" {
					method = "POST"
				}
				if action == "update" || action == "archive" || action == "restore" {
					method = "PATCH"
				}
				if action == "delete" {
					method = "DELETE"
				}
				body := `{"displayName":"Condition","enumOptions":[],"customAssetTypeIds":[]}`
				if action == "create" {
					body = `{"key":"condition","displayName":"Condition","type":"text","applicability":"all_assets","enumOptions":[],"customAssetTypeIds":[]}`
				}
				accepted := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
						w.WriteHeader(403)
						return
					}
					accepted++
					if r.Method != method {
						t.Error("wrong method")
					}
					if action == "list" && (r.URL.Query().Get("lifecycleState") != "all" || r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "next") {
						t.Error("query lost")
					}
					if action == "create" || action == "update" {
						b, _ := io.ReadAll(r.Body)
						if string(b) != body {
							t.Errorf("body changed: %s", b)
						}
					}
					if action == "delete" {
						w.WriteHeader(204)
						return
					}
					data := `{"id":"field","tenantId":"home","inventoryId":"tools","scope":"inventory","key":"condition","displayName":"Condition","type":"enum","applicability":"all_assets","enumOptions":[],"customAssetTypeIds":null,"lifecycleState":"active"}`
					if action == "list" {
						data = "[" + data + "]"
					}
					io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
				}))
				defer server.Close()
				for _, a := range []struct {
					token, tenant, inventory string
					allowed                  bool
				}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"viewer", "home", "tools", false}, {"owner", "other", "tools", false}, {"owner", "home", "other", false}} {
					if level == ports.HouseholdDefinition && a.inventory == "other" {
						continue
					}
					api, _ := New(server.URL, a.token, server.Client())
					scope := ports.DefinitionScope{Level: level, Scope: ports.Scope{Tenant: a.tenant, Inventory: a.inventory}}
					var result any
					var err error
					switch action {
					case "list":
						result, err = api.FieldDefinitions(context.Background(), scope, ports.Page{Limit: 2, Cursor: "next"}, "all")
					case "show":
						result, err = api.FieldDefinition(context.Background(), scope, "type")
					case "delete":
						err = api.DeleteFieldDefinition(context.Background(), scope, "type")
					default:
						result, err = api.ChangeFieldDefinition(context.Background(), scope, "type", ports.FieldAction(action), []byte(body))
					}
					if (err == nil) != a.allowed {
						t.Fatalf("scope: %+v %v", a, err)
					}
					if a.allowed && result != nil {
						b, _ := json.Marshal(result)
						for _, want := range []string{`"enumOptions":[]`, `"customAssetTypeIds":null`, `"requestId":"trace"`} {
							if !strings.Contains(string(b), want) {
								t.Fatalf("lost %s in %s", want, b)
							}
						}
					}
				}
				if accepted != 1 {
					t.Fatalf("accepted=%d", accepted)
				}
			})
		}
	}
}

func TestFieldDefinitionScopeCannotFallBackToHousehold(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }))
	defer server.Close()
	api, _ := New(server.URL, "owner", server.Client())
	for _, scope := range []ports.DefinitionScope{{}, {Level: "unknown", Scope: ports.Scope{Tenant: "home"}}, {Level: ports.InventoryDefinition, Scope: ports.Scope{Tenant: "home"}}, {Level: ports.HouseholdDefinition}} {
		if _, err := api.FieldDefinitions(context.Background(), scope, ports.Page{Limit: 2}, ""); err == nil {
			t.Fatal("list accepted missing scope")
		}
		if _, err := api.FieldDefinition(context.Background(), scope, "type"); err == nil {
			t.Fatal("show accepted missing scope")
		}
		if _, err := api.ChangeFieldDefinition(context.Background(), scope, "type", ports.UpdateField, []byte(`{}`)); err == nil {
			t.Fatal("write accepted missing scope")
		}
		if err := api.DeleteFieldDefinition(context.Background(), scope, "type"); err == nil {
			t.Fatal("delete accepted missing scope")
		}
	}
	if calls != 0 {
		t.Fatalf("invalid scope sent %d requests", calls)
	}
}
