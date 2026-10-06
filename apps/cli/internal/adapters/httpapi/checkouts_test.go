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

func TestCheckoutContractAndIsolation(t *testing.T) {
	body := `{"id":"checkout","assetId":"asset","tenantId":"home","inventoryId":"tools","state":"returned","checkedOutAt":"start","checkedOutByPrincipalId":"owner","checkoutDetails":"","returnedAt":"end","returnedByPrincipalId":"owner","returnDetails":"","createdAt":"created","updatedAt":"updated","undoableOperationId":"undo"}`
	for _, action := range []string{"history", "checkout", "return", "return-details"} {
		t.Run(action, func(t *testing.T) {
			path := "/tenants/home/inventories/tools/assets/asset/" + action
			method := "POST"
			if action == "history" {
				path = "/tenants/home/inventories/tools/assets/asset/checkouts"
				method = "GET"
			}
			if action == "return-details" {
				path = "/tenants/home/inventories/tools/assets/asset/checkouts/checkout/return-details"
				method = "PATCH"
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
				data := body
				if action == "history" {
					if r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "next" {
						t.Error("pagination lost")
					}
					data = "[" + body + "]"
				} else {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"details":""}` {
						t.Error("empty notes lost")
					}
				}
				io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
			}))
			defer server.Close()
			for _, attempt := range []struct {
				token, tenant, inventory string
				allowed                  bool
			}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"other", "home", "tools", false}, {"owner", "other", "tools", false}, {"owner", "home", "other", false}} {
				api, _ := New(server.URL, attempt.token, server.Client())
				scope := ports.Scope{Tenant: attempt.tenant, Inventory: attempt.inventory}
				var result any
				var err error
				if action == "history" {
					result, err = api.Checkouts(context.Background(), scope, "asset", ports.Page{Limit: 2, Cursor: "next"})
				} else {
					result, err = api.ChangeCheckout(context.Background(), scope, "asset", "checkout", ports.CheckoutAction(action), []byte(`{"details":""}`))
				}
				if (err == nil) != attempt.allowed {
					t.Fatalf("scope: %+v %v", attempt, err)
				}
				if attempt.allowed {
					encoded, _ := json.Marshal(result)
					for _, field := range []string{`"checkoutDetails":""`, `"returnDetails":""`, `"undoableOperationId":"undo"`, `"requestId":"trace"`, `"$schema":"schema"`} {
						if !strings.Contains(string(encoded), field) {
							t.Fatalf("lost field %s in %s", field, encoded)
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
