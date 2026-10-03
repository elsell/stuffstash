package httpserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/pairingcrypto"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPrintPairingHTTPBindsKeyScopeAndCredentialWithoutHumanAccess(t *testing.T) {
	runPrintPairingHTTP(t)
}
func runPrintPairingHTTP(t *testing.T) {
	const tid = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const iid = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	store := memory.NewStore()
	auth := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: tid, name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: iid, tenantID: tid, name: "Garage", owner: "owner"}}}, store, auth).WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintConnectors(store, auth, pairingcrypto.Secrets{}, printregistry.ConnectorPolicy{PublicWebBaseURL: "https://example.test", PairingLifetime: 10 * time.Minute, CredentialLifetime: 24 * time.Hour, ActivationLifetime: time.Minute, AuthorizationTimeout: time.Second, ReportMaxAge: time.Minute})
	server := NewServer(":0", application)
	printer, _, err := application.PrinterRegistry().Register(context.Background(), printregistry.RegisterPrinter{Actor: printregistry.Actor{Principal: identity.Principal{ID: "owner"}, Scope: printing.Scope{TenantID: tid, InventoryID: iid}}, RequestKey: "p", Name: "Garage", AdapterID: "brother-ql800", PresetID: "brother-ql800-29x90", PresetVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	start := performRequest(server, http.MethodPost, "/print-connector-pairings", "", map[string]any{"name": "Garage host", "publicKey": base64.StdEncoding.EncodeToString(pub), "candidates": []map[string]string{{"id": "usb-one", "name": "Brother", "adapterId": "brother-ql800", "deviceId": "protected-usb-path"}}})
	if start.Code != http.StatusCreated {
		t.Fatalf("pair create %d %s", start.Code, start.Body.String())
	}
	var pair struct {
		Data struct{ ID, PollToken, UserCode string }
	}
	if err := json.Unmarshal(start.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	path := "/print-connector-pairings/" + pair.Data.ID
	body := map[string]any{"userCode": pair.Data.UserCode, "tenantId": tid, "inventoryId": iid, "bindings": []map[string]string{{"candidateId": "usb-one", "printerId": string(printer.ID)}}}
	for _, token := range []string{"", "Bearer dev:other"} {
		r := performRequest(server, http.MethodPost, path+"/approval", token, body)
		if r.Code != http.StatusUnauthorized && r.Code != http.StatusForbidden {
			t.Fatalf("approval unauthorized %d", r.Code)
		}
	}

	poll := performRequestWithHeaders(server, http.MethodGet, path, "", map[string]string{"X-Pairing-Token": pair.Data.PollToken}, nil)
	requireStatus(t, poll, http.StatusOK)
	badPoll := performRequestWithHeaders(server, http.MethodGet, path, "", map[string]string{"X-Pairing-Token": "wrong"}, nil)
	requireStatus(t, badPoll, http.StatusUnauthorized)
	review := performRequest(server, http.MethodPost, path+"/review", "Bearer dev:owner", map[string]any{"userCode": pair.Data.UserCode, "tenantId": tid, "inventoryId": iid})
	requireStatus(t, review, http.StatusOK)
	if strings.Contains(review.Body.String(), "protected-usb-path") {
		t.Fatal("protected device identity leaked to approval preview")
	}
	badReview := performRequest(server, http.MethodPost, path+"/review", "Bearer dev:other", map[string]any{"userCode": pair.Data.UserCode, "tenantId": tid, "inventoryId": iid})
	requireStatus(t, badReview, http.StatusForbidden)
	badCreate := performRequest(server, http.MethodPost, "/print-connector-pairings", "", map[string]any{"name": "Malformed", "publicKey": "", "candidates": []any{}})
	requireStatus(t, badCreate, http.StatusUnprocessableEntity)
	auth.SetPrintingAvailable(false)
	approve := performRequest(server, http.MethodPost, path+"/approval", "Bearer dev:owner", body)
	if approve.Code != http.StatusOK {
		t.Fatalf("approve %d %s", approve.Code, approve.Body.String())
	}
	proof := ed25519.Sign(key, printing.PairingProofMessage(printing.PairingID(pair.Data.ID), pair.Data.PollToken))
	headers := map[string]string{"X-Pairing-Token": pair.Data.PollToken}
	bad := performRequestWithHeaders(server, http.MethodPost, path+"/credential", "", headers, map[string]string{"signature": base64.StdEncoding.EncodeToString(make([]byte, 64))})
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("forged proof %d", bad.Code)
	}
	pending := performRequestWithHeaders(server, http.MethodPost, path+"/credential", "", headers, map[string]string{"signature": base64.StdEncoding.EncodeToString(proof)})
	if pending.Code != http.StatusConflict {
		t.Fatalf("pending grants issued credential %d", pending.Code)
	}
	auth.SetPrintingAvailable(true)
	if err := application.PrintConnectors().DrainAuthorization(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	exchange := performRequestWithHeaders(server, http.MethodPost, path+"/credential", "", headers, map[string]string{"signature": base64.StdEncoding.EncodeToString(proof)})
	if exchange.Code != http.StatusOK {
		t.Fatalf("exchange %d %s", exchange.Code, exchange.Body.String())
	}
	var credential struct {
		Data struct{ Credential, ConnectorID string }
	}
	if err := json.Unmarshal(exchange.Body.Bytes(), &credential); err != nil {
		t.Fatal(err)
	}
	replay := performRequestWithHeaders(server, http.MethodPost, path+"/credential", "", headers, map[string]string{"signature": base64.StdEncoding.EncodeToString(proof)})
	if replay.Code < 400 {
		t.Fatal("credential exchange replayed")
	}
	heartbeat := performRequest(server, http.MethodPost, "/print-consumer/heartbeat", "Bearer "+credential.Data.Credential, map[string]string{"sessionId": "session-one"})
	if heartbeat.Code != http.StatusOK {
		t.Fatalf("heartbeat %d %s", heartbeat.Code, heartbeat.Body.String())
	}

	configs := performRequest(server, http.MethodGet, "/print-consumer/printers", "Bearer "+credential.Data.Credential, nil)
	if configs.Code != http.StatusOK {
		t.Fatalf("consumer configuration %d %s", configs.Code, configs.Body.String())
	}
	report := performRequest(server, http.MethodPost, "/print-consumer/printer-reports", "Bearer "+credential.Data.Credential, map[string]string{"printerId": string(printer.ID), "state": "ready"})
	if report.Code != http.StatusOK {
		t.Fatalf("report %d %s", report.Code, report.Body.String())
	}
	registered := performRequest(server, http.MethodGet, "/tenants/"+tid+"/inventories/"+iid+"/printers/"+string(printer.ID), "Bearer dev:owner", nil)
	var readiness struct{ Data struct{ Readiness string } }
	if err := json.Unmarshal(registered.Body.Bytes(), &readiness); err != nil || readiness.Data.Readiness != "ready" {
		t.Fatalf("ready printer not visible %s", registered.Body.String())
	}
	forgedReport := performRequest(server, http.MethodPost, "/print-consumer/printer-reports", "Bearer "+credential.Data.Credential, map[string]string{"printerId": "unassigned", "state": "ready"})
	if forgedReport.Code != http.StatusForbidden {
		t.Fatalf("unassigned printer report %d", forgedReport.Code)
	}
	human := performRequest(server, http.MethodGet, "/tenants/"+tid+"/inventories/"+iid+"/assets", "Bearer "+credential.Data.Credential, nil)
	if human.Code != http.StatusUnauthorized {
		t.Fatalf("worker accessed human inventory %d", human.Code)
	}

	collection := "/tenants/" + tid + "/inventories/" + iid + "/print-connectors"
	detail := collection + "/" + credential.Data.ConnectorID
	for _, resource := range []string{collection, detail} {
		allowed := performRequest(server, http.MethodGet, resource, "Bearer dev:owner", nil)
		requireStatus(t, allowed, http.StatusOK)
		if strings.Contains(allowed.Body.String(), credential.Data.Credential) || strings.Contains(allowed.Body.String(), "protected-usb-path") {
			t.Fatal("connector secret leaked to human read")
		}
		forbidden := performRequest(server, http.MethodGet, resource, "Bearer dev:other", nil)
		requireStatus(t, forbidden, http.StatusForbidden)
	}
	wrongUpdate := performRequest(server, http.MethodPatch, detail, "Bearer dev:other", map[string]any{"generation": 1, "revoked": true})
	requireStatus(t, wrongUpdate, http.StatusForbidden)
	badHeartbeat := performRequest(server, http.MethodPost, "/print-consumer/heartbeat", "Bearer dev:owner", map[string]string{"sessionId": "session-one"})
	requireStatus(t, badHeartbeat, http.StatusUnauthorized)
	badConfig := performRequest(server, http.MethodGet, "/print-consumer/printers", "Bearer dev:owner", nil)
	requireStatus(t, badConfig, http.StatusUnauthorized)
	revoke := performRequest(server, http.MethodPatch, "/tenants/"+tid+"/inventories/"+iid+"/print-connectors/"+credential.Data.ConnectorID, "Bearer dev:owner", map[string]any{"generation": 1, "revoked": true})
	if revoke.Code != http.StatusOK {
		t.Fatalf("revoke %d %s", revoke.Code, revoke.Body.String())
	}
	denied := performRequest(server, http.MethodPost, "/print-consumer/heartbeat", "Bearer "+credential.Data.Credential, map[string]string{"sessionId": "session-one"})
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("revoked credential accepted %d", denied.Code)
	}

}

func coverPrintPairingScenarios(t *testing.T, coverage executedScenarioCoverage, _ bool) {
	t.Helper()
	// This integrated scenario executes and asserts both legitimate and adversarial
	// requests for each operation before recording that operation as covered.
	runPrintPairingHTTP(t)
	for _, operation := range []string{
		"POST /print-connector-pairings", "GET /print-connector-pairings/{pairingId}",
		"POST /print-connector-pairings/{pairingId}/review", "POST /print-connector-pairings/{pairingId}/approval", "POST /print-connector-pairings/{pairingId}/credential",
		"POST /print-consumer/heartbeat", "GET /print-consumer/printers", "POST /print-consumer/printer-reports",
		"GET /tenants/{tenantId}/inventories/{inventoryId}/print-connectors", "GET /tenants/{tenantId}/inventories/{inventoryId}/print-connectors/{connectorId}", "PATCH /tenants/{tenantId}/inventories/{inventoryId}/print-connectors/{connectorId}",
	} {
		coverage.operation[operation] = struct{}{}
	}
}
