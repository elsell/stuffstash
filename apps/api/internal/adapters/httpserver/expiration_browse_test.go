package httpserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestExpirationBrowseIncludesPersonalStatus(t *testing.T) {
	server, _ := notificationHTTPFixture(t)
	for _, path := range []string{"/tenants/home/inventories/main/assets", "/tenants/home/search/assets?inventoryId=main&q=Bottle"} {
		response := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
		requireStatus(t, response, http.StatusOK)
		var body struct {
			Data []struct {
				ExpirationContext *struct{ State string }
				Asset             *struct{ ExpirationContext *struct{ State string } }
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Data) != 1 {
			t.Fatalf("expected one dated asset: %s", response.Body.String())
		}
		context := body.Data[0].ExpirationContext
		if body.Data[0].Asset != nil {
			context = body.Data[0].Asset.ExpirationContext
		}
		if context == nil || context.State != "expired" {
			t.Fatalf("missing browse expiration status: %s", response.Body.String())
		}
	}
}
