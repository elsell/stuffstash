package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"net/http"
	"testing"
)

func TestNotificationInboxHTTPAccessAndReadState(t *testing.T) {
	server, store := notificationHTTPFixture(t)
	const base = "/tenants/home/inventories/main/notifications"
	for _, request := range []struct{ method, path string }{{http.MethodGet, base}, {http.MethodGet, base + "/notice"}, {http.MethodPut, base + "/notice/read"}} {
		for _, auth := range []string{"", "Bearer malformed", "Bearer dev:outsider"} {
			response := performRequest(server, request.method, request.path, auth, nil)
			if response.Code != http.StatusUnauthorized && response.Code != http.StatusForbidden {
				t.Fatalf("unauthorized inbox %d %s", response.Code, response.Body.String())
			}
		}
	}
	response := performRequest(server, http.MethodGet, base, "Bearer dev:owner", nil)
	requireStatus(t, response, http.StatusOK)
	var body struct {
		Data []struct{ ID, AssetID, Title string }
		Meta struct{ Pagination struct{ HasMore bool } }
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 1 || body.Data[0].AssetID != "bottle" || body.Data[0].Title != "Bottle" || body.Meta.Pagination.HasMore {
		t.Fatalf("list %+v", body)
	}
	requireStatus(t, performRequest(server, http.MethodGet, base+"/notice", "Bearer dev:viewer", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodPut, base+"/notice/read", "Bearer dev:viewer", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/other/inventories/main/notifications", "Bearer dev:owner", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodPut, base+"/notice/read", "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPut, base+"/notice/read", "Bearer dev:owner", nil), http.StatusOK)
	response = performRequest(server, http.MethodGet, base+"?unreadOnly=true", "Bearer dev:owner", nil)
	requireStatus(t, response, http.StatusOK)
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 0 {
		t.Fatal("read state not reflected")
	}
	item, _, _ := store.AssetByID(context.Background(), "home", "main", "bottle")
	item.Expiration = expirationdate.Date{}
	if err := store.UpdateAsset(context.Background(), item, []audit.Record{{ID: "clear-expiration"}}, nil); err != nil {
		t.Fatal(err)
	}
	requireStatus(t, performRequest(server, http.MethodGet, base+"/notice", "Bearer dev:owner", nil), http.StatusNotFound)
}
