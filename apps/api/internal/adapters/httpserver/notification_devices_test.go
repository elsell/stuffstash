package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"net/http"
	"strings"
	"testing"
)

func TestNotificationDevicesHTTPAuthenticationOwnershipAndCleanup(t *testing.T) {
	store := memory.NewStore()
	auth := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: "main", tenantID: "home", name: "Main", owner: "owner"}}}, store, auth)
	viewer := identity.Principal{ID: "viewer"}
	if err := auth.GrantInventoryViewer(context.Background(), viewer, "home", "main"); err != nil {
		t.Fatal(err)
	}
	server := NewServer(":0", application)
	base := "/tenants/home/inventories/main/notification-devices"
	token := strings.Repeat("ab", 32)
	body := map[string]any{"installationId": "phone", "transport": "apns", "token": token, "revision": 0}
	for _, auth := range []string{"", "Bearer malformed"} {
		requireStatus(t, performRequest(server, http.MethodPost, base, auth, body), http.StatusUnauthorized)
	}
	requireStatus(t, performRequest(server, http.MethodPost, base, "Bearer dev:outsider", body), http.StatusForbidden)
	response := performRequest(server, http.MethodPost, base, "Bearer dev:viewer", body)
	requireStatus(t, response, http.StatusOK)
	if strings.Contains(response.Body.String(), token) {
		t.Fatal("token exposed")
	}
	var decoded struct {
		Data struct {
			ID       string
			Revision int64
			Active   bool
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Data.ID == "" || decoded.Data.Revision != 1 || !decoded.Data.Active {
		t.Fatal("invalid registration response")
	}
	device := decoded.Data.ID
	body["token"] = strings.ToUpper(token)
	requireStatus(t, performRequest(server, http.MethodPost, base, "Bearer dev:owner", body), http.StatusConflict)
	body["token"] = token
	for _, auth := range []string{"", "Bearer malformed"} {
		requireStatus(t, performRequest(server, http.MethodGet, base+"/by-installation/phone", auth, nil), http.StatusUnauthorized)
		requireStatus(t, performRequest(server, http.MethodDelete, base+"/"+device+"?revision=1", auth, nil), http.StatusUnauthorized)
	}
	requireStatus(t, performRequest(server, http.MethodDelete, base+"/"+device+"?revision=1", "Bearer dev:owner", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodGet, base+"/by-installation/phone", "Bearer dev:owner", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/other/inventories/main/notification-devices/by-installation/phone", "Bearer dev:viewer", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodDelete, base+"/"+device+"?revision=9", "Bearer dev:viewer", nil), http.StatusConflict)
	if err := auth.RevokeInventoryViewer(context.Background(), viewer, "home", "main"); err != nil {
		t.Fatal(err)
	}
	requireStatus(t, performRequest(server, http.MethodPost, base, "Bearer dev:viewer", body), http.StatusForbidden)
	requireStatus(t, performRequest(server, http.MethodGet, base+"/by-installation/phone", "Bearer dev:viewer", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodDelete, base+"/"+device+"?revision=1", "Bearer dev:viewer", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPost, base, "Bearer dev:owner", body), http.StatusOK)
	body["token"] = "not-hex"
	requireStatus(t, performRequest(server, http.MethodPost, base, "Bearer dev:owner", body), http.StatusBadRequest)
	body["token"] = strings.Repeat("secret", 700)
	invalid := performRequest(server, http.MethodPost, base, "Bearer dev:owner", body)
	requireStatus(t, invalid, http.StatusBadRequest)
	if strings.Contains(invalid.Body.String(), "secret") {
		t.Fatal("validation reflected token")
	}

}
