package httpserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestNotificationInboxBatchHTTPIsolation(t *testing.T) {
	server, _ := notificationHTTPFixture(t)
	const base = "/tenants/home/inventories/main/notifications"
	for _, operation := range []struct{ method, suffix string }{{http.MethodGet, "/unread-count"}, {http.MethodPut, "/read-all"}} {
		for _, token := range []string{"", "Bearer malformed", "Bearer dev:outsider"} {
			response := performRequest(server, operation.method, base+operation.suffix, token, nil)
			if response.Code != http.StatusUnauthorized && response.Code != http.StatusForbidden {
				t.Fatalf("unauthorized batch %d", response.Code)
			}
		}
		requireStatus(t, performRequest(server, operation.method, "/tenants/other/inventories/main/notifications"+operation.suffix, "Bearer dev:owner", nil), http.StatusNotFound)
		requireStatus(t, performRequest(server, operation.method, base+operation.suffix, "Bearer dev:viewer", nil), http.StatusOK)
	}
	count := func(want int) {
		t.Helper()
		response := performRequest(server, http.MethodGet, base+"/unread-count", "Bearer dev:owner", nil)
		requireStatus(t, response, http.StatusOK)
		var body struct {
			Data struct{ Count int }
			Meta struct{ Pagination struct{ HasMore bool } }
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Data.Count != want || body.Meta.Pagination.HasMore {
			t.Fatalf("count %+v", body)
		}
	}
	count(1) // Another member's mark-all must not mark the owner's notification.
	for i := 0; i < 2; i++ {
		response := performRequest(server, http.MethodPut, base+"/read-all", "Bearer dev:owner", nil)
		requireStatus(t, response, http.StatusOK)
		var body struct{ Data struct{ Complete bool } }
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if !body.Data.Complete {
			t.Fatal("final page not complete")
		}
	}
	count(0)
}
