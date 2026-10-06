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

func TestAssetReadsPreserveCompleteFieldsAndFilters(t *testing.T) {
	assetJSON := `{"id":"asset","tenantId":"home","inventoryId":"garage","kind":"item","title":"Drill","description":"Cordless","parentAssetId":"","customAssetTypeId":"type","customFields":{"serial":9007199254740993,"working":false},"expiration":{"date":"2027-01","precision":"month"},"expirationContext":{"state":"upcoming","trackingEnabled":true,"advanceDays":30,"timezone":"UTC"},"tags":[],"currentCheckout":{"id":"checkout","state":"checked_out","checkedOutAt":"today","checkedOutByPrincipalId":"alice","checkedOutByPrincipal":{"id":"alice","email":"alice@example.test"}},"primaryPhoto":{"id":"photo","fileName":"drill.jpg","contentType":"image/jpeg","sizeBytes":12,"thumbnails":{"small":"s","medium":"m","large":"l"}},"printJobId":"","undoableOperationId":"undo","lifecycleState":"active","createdAt":"created","updatedAt":"updated"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		data := assetJSON
		if strings.HasSuffix(r.URL.Path, "/assets") {
			if r.URL.Query().Get("lifecycleState") != "all" || r.URL.Query().Get("sort") != "updated_desc" {
				t.Errorf("filters lost: %s", r.URL)
			}
			data = "[" + data + "]"
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
	}))
	defer server.Close()
	api, _ := New(server.URL, "owner", server.Client())
	result, err := api.Asset(context.Background(), ports.Scope{Tenant: "home", Inventory: "garage"}, "asset")
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.Marshal(result)
	var expected, got map[string]json.RawMessage
	json.Unmarshal([]byte(assetJSON), &expected)
	data, _ := json.Marshal(result.Data)
	json.Unmarshal(data, &got)
	for key, value := range expected {
		if string(got[key]) != string(value) {
			var a, b any
			da := json.NewDecoder(strings.NewReader(string(value)))
			da.UseNumber()
			da.Decode(&a)
			db := json.NewDecoder(strings.NewReader(string(got[key])))
			db.UseNumber()
			db.Decode(&b)
			left, _ := json.Marshal(a)
			right, _ := json.Marshal(b)
			if string(left) != string(right) {
				t.Errorf("field %s changed: got=%s want=%s", key, got[key], value)
			}
		}
	}
	if !strings.Contains(string(bytes), `"requestId":"trace"`) {
		t.Fatal("metadata lost")
	}
	list, err := api.Assets(context.Background(), ports.Scope{Tenant: "home", Inventory: "garage"}, ports.AssetQuery{Page: ports.Page{Limit: 50}, Lifecycle: "all", Sort: "updated_desc"})
	if err != nil || len(list.Data) != 1 {
		t.Fatalf("list failed: %+v %v", list, err)
	}
	listData, err := json.Marshal(list.Data[0])
	if err != nil || string(listData) != string(data) {
		t.Fatalf("list asset differs from detail: %s, detail: %s, error: %v", listData, data, err)
	}
}
