package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"net/http"
	"testing"
)

func TestPrinterRegistrationEnforcesRolesScopeIdempotencyAndRevisions(t *testing.T) {
	const tid = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const iid = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: tid, name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: iid, tenantID: tid, name: "Garage", owner: "owner"}}}, store, authorizer).WithPrinterRegistry(store, printingprofiles.Catalog{})
	if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, tenant.ID(tid), inventory.InventoryID(iid)); err != nil {
		t.Fatal(err)
	}
	server := NewServer(":0", application)
	path := "/tenants/" + tid + "/inventories/" + iid + "/printers"
	body := map[string]any{"name": "Garage Brother", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1}
	for _, test := range []struct {
		token  string
		status int
	}{{"", http.StatusUnauthorized}, {"Bearer malformed", http.StatusUnauthorized}, {"Bearer dev:other", http.StatusForbidden}, {"Bearer dev:viewer", http.StatusForbidden}} {
		r := performRequestWithHeaders(server, http.MethodPost, path, test.token, map[string]string{"Idempotency-Key": "request-one"}, body)
		if r.Code != test.status {
			t.Fatalf("unauthorized registration %d: %s", r.Code, r.Body.String())
		}
	}
	create := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:owner", map[string]string{"Idempotency-Key": "request-one"}, body)
	if create.Code != http.StatusCreated {
		t.Fatalf("create %d: %s", create.Code, create.Body.String())
	}
	var created struct {
		Data struct {
			ID       string `json:"id"`
			Revision uint64 `json:"revision"`
		}
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	replay := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:owner", map[string]string{"Idempotency-Key": "request-one"}, body)
	if replay.Code != http.StatusOK {
		t.Fatalf("replay %d: %s", replay.Code, replay.Body.String())
	}
	body["name"] = "Different"
	collision := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:owner", map[string]string{"Idempotency-Key": "request-one"}, body)
	if collision.Code != http.StatusConflict {
		t.Fatalf("key reused %d", collision.Code)
	}
	detail := path + "/" + created.Data.ID
	read := performRequest(server, http.MethodGet, detail, "Bearer dev:viewer", nil)
	if read.Code != http.StatusOK {
		t.Fatalf("viewer read %d: %s", read.Code, read.Body.String())
	}
	wrongScope := performRequest(server, http.MethodGet, "/tenants/"+tid+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAX/printers/"+created.Data.ID, "Bearer dev:owner", nil)
	if wrongScope.Code != http.StatusNotFound {
		t.Fatalf("cross inventory %d", wrongScope.Code)
	}
	update := performRequest(server, http.MethodPatch, detail, "Bearer dev:owner", map[string]any{"revision": created.Data.Revision, "name": "Renamed"})
	if update.Code != http.StatusOK {
		t.Fatalf("update %d: %s", update.Code, update.Body.String())
	}
	stale := performRequest(server, http.MethodPatch, detail, "Bearer dev:owner", map[string]any{"revision": created.Data.Revision, "name": "Stale"})
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale update %d", stale.Code)
	}
	scope := printing.Scope{TenantID: tid, InventoryID: iid}
	reserved, err := store.UpdatePrinter(context.Background(), scope, printing.PrinterID(created.Data.ID), 2, func(p *printing.Printer) error {
		p.ActiveJobID = "started-job"
		p.ReservationState = "printing"
		p.MediaFingerprint = "previous-media"
		p.Revision++
		return nil
	}, func(printing.Printer) (audit.Record, error) { return audit.Record{ID: "test-queue-start"}, nil })
	if err != nil {
		t.Fatal(err)
	}
	blocked := performRequest(server, http.MethodPatch, detail, "Bearer dev:owner", map[string]any{"revision": reserved.Revision, "presetId": "brother-ql800-29x90", "presetVersion": 1})
	if blocked.Code != http.StatusConflict {
		t.Fatalf("changed started job media: %d %s", blocked.Code, blocked.Body.String())
	}
	retire := performRequest(server, http.MethodPatch, detail, "Bearer dev:owner", map[string]any{"revision": reserved.Revision, "retired": true})
	if retire.Code != http.StatusOK {
		t.Fatalf("retire %d %s", retire.Code, retire.Body.String())
	}
	retained, err := store.GetPrinter(context.Background(), scope, printing.PrinterID(created.Data.ID))
	if err != nil || !retained.Retired || retained.ActiveJobID != "started-job" || retained.ReservationState != "printing" {
		t.Fatalf("retirement lost active output evidence: %+v %v", retained, err)
	}

}

func coverPrinterRegistryScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	const tid = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const iid = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: tid, name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: iid, tenantID: tid, name: "Garage", owner: "owner"}}}, store, authorizer).WithPrinterRegistry(store, printingprofiles.Catalog{})
	server := NewServer(":0", application)
	path := "/tenants/" + tid + "/inventories/" + iid + "/printers"
	template := "/tenants/{tenantId}/inventories/{inventoryId}/printers"
	body := map[string]any{"name": "Garage Brother", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1}
	token := "Bearer dev:owner"
	status := http.StatusOK
	postStatus := http.StatusCreated
	if adversarial {
		token = "Bearer dev:other"
		status = http.StatusForbidden
		postStatus = status
	}
	response := performRequestWithHeaders(server, http.MethodPost, path, token, map[string]string{"Idempotency-Key": "scenario-key"}, body)
	requireStatus(t, response, postStatus)
	coverage.operation[http.MethodPost+" "+template] = struct{}{}
	printerID := "unrelated-printer"
	if !adversarial {
		var created struct {
			Data struct {
				ID string `json:"id"`
			}
		}
		decodeBody(t, response, &created)
		printerID = created.Data.ID
	}
	coverage.request(t, server, http.MethodGet, template, path, token, nil, status)
	coverage.request(t, server, http.MethodGet, template+"/{printerId}", path+"/"+printerID, token, nil, status)
	coverage.request(t, server, http.MethodPatch, template+"/{printerId}", path+"/"+printerID, token, map[string]any{"revision": 1, "name": "Renamed"}, status)
}
