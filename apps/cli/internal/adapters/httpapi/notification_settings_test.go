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

func TestNotificationSettingsTransport(t *testing.T) {
	for _, tc := range []struct{ action, method, path, body string }{
		{"show", "GET", "notification-preferences", ""},
		{"update", "PUT", "notification-preferences", `{"revision":9007199254740993,"defaults":{"enabled":false,"upcoming":false,"expired":true,"advanceDays":0},"timezone":"UTC","pushEnabled":false}`},
		{"initialize", "POST", "notification-preferences/initialize", `{"timezone":"UTC"}`},
		{"override", "PUT", "notification-preferences/types/type", `{"revision":9007199254740993,"settings":{"enabled":false,"upcoming":true,"expired":false,"advanceDays":0}}`},
		{"remove-override", "DELETE", "notification-preferences/types/type", ""},
		{"register-device", "POST", "notification-devices", `{"installationId":"install","transport":"apns","token":"secret-device-token","revision":9007199254740993}`},
		{"show-device", "GET", "notification-devices/by-installation/install", ""},
		{"remove-device", "DELETE", "notification-devices/device", ""},
	} {
		t.Run(tc.action, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/tenants/home/inventories/tools/"+tc.path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != tc.method {
					t.Error("method changed")
				}
				b, _ := io.ReadAll(r.Body)
				if string(b) != tc.body {
					t.Errorf("body changed: %s", b)
				}
				if tc.method == "DELETE" && r.URL.Query().Get("revision") != "9007199254740993" {
					t.Error("revision lost")
				}
				data := `{"revision":9007199254740993,"defaults":{"enabled":false,"upcoming":true,"expired":false,"advanceDays":0},"timezone":"UTC","pushEnabled":false,"overrides":[{"customAssetTypeId":"type","settings":{"enabled":true,"upcoming":false,"expired":true,"advanceDays":42}}]}`
				if strings.Contains(tc.action, "device") {
					data = `{"id":"device","installationId":"install","transport":"apns","revision":9007199254740993,"active":false}`
				}
				io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			for _, a := range []struct {
				token, tenant, inventory string
				allowed                  bool
			}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"other", "home", "tools", false}, {"owner", "other", "tools", false}, {"owner", "home", "other", false}} {
				api, _ := New(server.URL, a.token, server.Client())
				s := ports.Scope{Tenant: a.tenant, Inventory: a.inventory}
				var result any
				var err error
				before := calls
				switch tc.action {
				case "show":
					result, err = api.NotificationPreferences(context.Background(), s)
				case "show-device":
					result, err = api.NotificationDevice(context.Background(), s, "install")
				case "register-device":
					result, err = api.RegisterNotificationDevice(context.Background(), s, []byte(tc.body))
				case "remove-device":
					result, err = api.RemoveNotificationDevice(context.Background(), s, "device", 9007199254740993)
				default:
					result, err = api.ChangeNotificationPreferences(context.Background(), s, ports.PreferenceAction(tc.action), "type", 9007199254740993, []byte(tc.body))
				}
				if calls != before+1 {
					t.Fatal("unexpected retries")
				}
				if !a.allowed {
					if err == nil {
						t.Fatal("unauthorized request succeeded")
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				b, _ := json.Marshal(result)
				text := string(b)
				if !strings.Contains(text, `"revision":9007199254740993`) || !strings.Contains(text, `"requestId":"trace"`) || strings.Contains(text, "secret-device-token") {
					t.Fatal(text)
				}
				if !strings.Contains(tc.action, "device") && (!strings.Contains(text, `"pushEnabled":false`) || !strings.Contains(text, `"advanceDays":42`)) {
					t.Fatal(text)
				}
			}
		})
	}
}
