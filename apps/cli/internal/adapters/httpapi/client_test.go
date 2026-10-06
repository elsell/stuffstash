package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestGeneratedTransportPreservesScopeNullMoveAndDenial(t *testing.T) {
	var gotPath, gotBody, gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotKey = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer allowed" {
			w.WriteHeader(403)
			io.WriteString(w, `{"error":{"code":"forbidden","message":"secret provider details"}}`)
			return
		}
		io.WriteString(w, `{"data":{"id":"asset","tenantId":"tenant","inventoryId":"inventory","kind":"item","title":"Box","description":"","customFields":{},"tags":[],"expiration":null,"lifecycleState":"active","createdAt":"","updatedAt":""},"meta":{}}`)
	}))
	defer server.Close()
	client, err := New(server.URL, "allowed", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.UpdateAsset(context.Background(), ports.Scope{Tenant: "tenant", Inventory: "inventory"}, "asset", ports.AssetChange{MoveToRoot: true}, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Data.ID != "asset" || gotPath != "/tenants/tenant/inventories/inventory/assets/asset" || !strings.Contains(gotBody, `"parentAssetId":null`) || gotKey != "" {
		t.Fatalf("incorrect wire behavior: %#v %s %s %s", result, gotPath, gotBody, gotKey)
	}
	denied, _ := New(server.URL, "denied", server.Client())
	_, err = denied.Asset(context.Background(), ports.Scope{Tenant: "tenant", Inventory: "inventory"}, "asset")
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe denial: %v", err)
	}
}
