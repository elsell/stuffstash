package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// configuredPrinterServer implements scoped, permission-checked optimistic
// updates while preserving fields absent from a PATCH.
type configuredPrinterServer struct {
	mu       sync.Mutex
	revision uint64
	name     string
	retired  bool
	preset   string
}

func (f *configuredPrinterServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	token := r.Header.Get("Authorization")
	if token != "Bearer owner" && token != "Bearer viewer" {
		w.WriteHeader(401)
		return
	}
	if r.URL.Path == "/tenants/home/inventories/garage/printer-profiles" && r.Method == "GET" {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{
			map[string]any{"adapterId": "brother-ql", "media": []any{map[string]any{"presetId": "brother-ql800-29x90", "version": 1}}},
			map[string]any{"adapterId": "other-adapter", "media": []any{map[string]any{"presetId": "other-stock", "version": 1}}},
		}, "meta": map[string]any{}})
		return
	}
	if r.URL.Path != "/tenants/home/inventories/garage/printers/brother" {
		w.WriteHeader(404)
		return
	}
	if r.Method == "PATCH" {
		if token != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		var body map[string]json.RawMessage
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			w.WriteHeader(400)
			return
		}
		var revision uint64
		var preset string
		var version int
		_ = json.Unmarshal(body["revision"], &revision)
		_ = json.Unmarshal(body["presetId"], &preset)
		_ = json.Unmarshal(body["presetVersion"], &version)
		if revision != f.revision {
			w.WriteHeader(409)
			return
		}
		if preset != "brother-ql800-29x90" || version != 1 {
			w.WriteHeader(422)
			return
		}
		if value, ok := body["name"]; ok {
			_ = json.Unmarshal(value, &f.name)
		}
		if value, ok := body["retired"]; ok {
			_ = json.Unmarshal(value, &f.retired)
		}
		f.preset = preset
		f.revision++
	} else if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "brother", "name": f.name, "retired": f.retired, "revision": f.revision, "adapterId": "brother-ql", "readiness": "unavailable", "media": map[string]any{"presetId": f.preset, "version": 1}}, "meta": map[string]any{}})
}
func TestConfigurePrinterUsesRevisionAndPreservesUnchangedRegistration(t *testing.T) {
	fake := &configuredPrinterServer{revision: 4, name: "Garage Brother", retired: true, preset: "old-profile"}
	server := httptest.NewServer(fake)
	defer server.Close()
	client, _ := New(server.URL, "owner", server.Client())
	ctx := context.Background()
	scope := ports.Scope{Tenant: "home", Inventory: "garage"}
	current, err := client.RegisteredPrinter(ctx, scope, "brother")
	if err != nil {
		t.Fatal(err)
	}
	preset := ports.PrinterMediaPreset{ID: "brother-ql800-29x90", Version: 1}
	presets, err := client.PrinterMediaPresets(ctx, scope, current.AdapterID)
	if err != nil || len(presets) != 1 || presets[0] != preset {
		t.Fatalf("wrong adapter catalog: %+v %v", presets, err)
	}

	changed, err := client.ConfigurePrinterMedia(ctx, scope, current.ID, current.Revision, preset)
	if err != nil || changed.Data.Name != current.Name || changed.Data.Retired != current.Retired || changed.Data.MediaPreset != preset.ID || changed.Data.Revision != 5 {
		t.Fatalf("update: %+v %v", changed, err)
	}
	if _, err = client.ConfigurePrinterMedia(ctx, scope, current.ID, current.Revision, preset); err == nil {
		t.Fatal("stale write succeeded")
	}
	for _, token := range []string{"viewer", "", "expired", "connector"} {
		denied, _ := New(server.URL, token, server.Client())
		if _, err = denied.ConfigurePrinterMedia(ctx, scope, current.ID, 5, preset); err == nil {
			t.Fatalf("%q unauthorized write succeeded", token)
		}
	}
	for _, wrong := range []ports.Scope{{Tenant: "other", Inventory: "garage"}, {Tenant: "home", Inventory: "other"}} {
		if _, err = client.ConfigurePrinterMedia(ctx, wrong, current.ID, 5, preset); err == nil {
			t.Fatal("cross-scope write succeeded")
		}
	}
	if _, err = client.ConfigurePrinterMedia(ctx, scope, current.ID, 5, ports.PrinterMediaPreset{ID: "unsupported", Version: 1}); err == nil {
		t.Fatal("unsupported stock succeeded")
	}
	final, err := client.RegisteredPrinter(ctx, scope, "brother")
	if err != nil || final.Revision != 5 || final.Name != current.Name || final.Retired != current.Retired {
		t.Fatalf("rejected update mutated printer: %+v %v", final, err)
	}
}
