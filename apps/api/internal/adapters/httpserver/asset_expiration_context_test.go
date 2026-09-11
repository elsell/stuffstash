package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssetDetailExpirationContextIsPersonalAndScoped(t *testing.T) {
	server, _ := notificationHTTPFixture(t)
	base := "/tenants/home/inventories/main"
	requireStatus(t, performRequest(server, http.MethodPost, base+"/notification-preferences/initialize", "Bearer dev:owner", map[string]any{"timezone": "America/New_York"}), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPut, base+"/notification-preferences", "Bearer dev:owner", map[string]any{"revision": 1, "timezone": "America/New_York", "pushEnabled": false, "defaults": map[string]any{"enabled": false, "upcoming": false, "expired": false, "advanceDays": 7}}), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPut, base+"/notification-preferences/types/medicine", "Bearer dev:owner", map[string]any{"revision": 2, "settings": map[string]any{"enabled": false, "upcoming": false, "expired": false, "advanceDays": 14}}), http.StatusOK)
	for _, principal := range []string{"owner", "viewer"} {
		response := performRequest(server, http.MethodGet, base+"/assets/bottle", "Bearer dev:"+principal, nil)
		requireStatus(t, response, http.StatusOK)
		var body struct {
			Data struct {
				Context *struct {
					State           string
					TrackingEnabled bool
					AdvanceDays     int
					Timezone        string
				} `json:"expirationContext"`
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		c := body.Data.Context
		if c == nil || c.State != "expired" || !c.TrackingEnabled {
			t.Fatalf("missing factual expiration context: %s", response.Body.String())
		}
		days, zone := 30, "UTC"
		if principal == "owner" {
			days, zone = 14, "America/New_York"
		}
		if c.AdvanceDays != days || c.Timezone != zone {
			t.Fatalf("personal settings mixed: %+v", c)
		}
	}
	requireStatus(t, performRequest(server, http.MethodPatch, base+"/custom-asset-types/medicine", "Bearer dev:owner", map[string]any{"expirationEnabled": false}), http.StatusOK)
	disabled := performRequest(server, http.MethodGet, base+"/assets/bottle", "Bearer dev:owner", nil)
	requireStatus(t, disabled, http.StatusOK)
	var disabledBody struct {
		Data struct {
			Context struct {
				TrackingEnabled bool
				State           string
			} `json:"expirationContext"`
		}
	}
	if err := json.Unmarshal(disabled.Body.Bytes(), &disabledBody); err != nil {
		t.Fatal(err)
	}
	if disabledBody.Data.Context.TrackingEnabled || disabledBody.Data.Context.State != "expired" {
		t.Fatal("disabled tracking lost the factual date state")
	}
	requireStatus(t, performRequest(server, http.MethodPatch, base+"/assets/bottle", "Bearer dev:owner", map[string]any{"expiration": nil}), http.StatusOK)
	cleared := performRequest(server, http.MethodGet, base+"/assets/bottle", "Bearer dev:owner", nil)
	requireStatus(t, cleared, http.StatusOK)
	if strings.Contains(cleared.Body.String(), "expirationContext") {
		t.Fatal("cleared date retained status")
	}

	for _, test := range []struct {
		path, auth string
		status     int
	}{
		{base + "/assets/bottle", "", http.StatusUnauthorized},
		{base + "/assets/bottle", "Bearer malformed", http.StatusUnauthorized},
		{base + "/assets/bottle", "Bearer dev:outsider", http.StatusForbidden},
		{"/tenants/other/inventories/main/assets/bottle", "Bearer dev:owner", http.StatusNotFound},
		{"/tenants/home/inventories/other/assets/bottle", "Bearer dev:owner", http.StatusNotFound},
	} {
		response := performRequest(server, http.MethodGet, test.path, test.auth, nil)
		if response.Code != test.status && !(test.status == http.StatusNotFound && response.Code == http.StatusForbidden) {
			t.Fatalf("boundary accepted: %d", response.Code)
		}
		if strings.Contains(response.Body.String(), "expirationContext") {
			t.Fatal("denied read leaked personal context")
		}
	}
}
