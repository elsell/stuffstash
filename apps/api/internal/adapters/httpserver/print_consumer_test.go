package httpserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/pairingcrypto"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func TestPrintConsumerOwnsAttemptAndReportsPhysicalEvidence(t *testing.T) {
	coverPrintConsumerScenarios(t, newExecutedScenarioCoverage("print consumer"))
}

func coverPrintConsumerScenarios(t *testing.T, coverage executedScenarioCoverage) {
	ctx := context.Background()
	clock := &labelTestClock{now: time.Now().UTC()}
	application, store, az := labelTestApplication(t, clock)
	application = application.WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintJobs(store, printingapp.JobConfig{MaxCopies: 20, MaxArtifactBytes: 1000000, ArtifactTTL: time.Hour, Lease: 5 * time.Second, ReadinessMaxAge: time.Second}).WithPrintConnectors(store, az, pairingcrypto.Secrets{}, printregistry.ConnectorPolicy{PublicWebBaseURL: "https://example.test", PairingLifetime: time.Minute, CredentialLifetime: time.Hour, ActivationLifetime: time.Minute, AuthorizationTimeout: time.Second, ReportMaxAge: time.Minute})
	actor := printregistry.Actor{Principal: identity.Principal{ID: "owner"}, Scope: printing.Scope{TenantID: labelTenant, InventoryID: labelInventory}}
	p, _, err := application.PrinterRegistry().Register(ctx, printregistry.RegisterPrinter{Actor: actor, Name: "Garage", RequestKey: "printer", AdapterID: "brother-ql800", PresetID: "brother-ql800-29x90", PresetVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := application.PrintConnectors().Begin(ctx, printregistry.BeginPairing{Name: "Host", PublicKey: pub, Candidates: []printing.PairingCandidate{{ID: "usb", Name: "Brother", AdapterID: "brother-ql800", DeviceID: "protected-device"}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = application.PrintConnectors().Approve(ctx, printregistry.ApprovePairing{Actor: actor, PairingID: pair.Pairing.ID, UserCode: pair.UserCode, Bindings: []printregistry.PairingBinding{{CandidateID: "usb", PrinterID: p.ID}}})
	if err != nil {
		t.Fatal(err)
	}
	credential, err := application.PrintConnectors().Exchange(ctx, pair.Pairing.ID, pair.PollToken, ed25519.Sign(key, printing.PairingProofMessage(pair.Pairing.ID, pair.PollToken)))
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(":0", application)
	token := "Bearer " + credential.Credential
	requireStatus(t, performRequest(server, "POST", "/print-consumer/heartbeat", token, map[string]any{"sessionId": "session-for-printing"}), 200)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/printer-reports", token, map[string]any{"printerId": p.ID, "state": "ready"}), 200)
	asset := performRequest(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", map[string]any{"title": "Tools", "kind": "container"})
	requireStatus(t, asset, 201)
	assetID := decodeAsset(t, asset).Data.ID
	enqueue := performRequestWithHeaders(server, "POST", labelPrefix+"/assets/"+assetID+"/print-jobs", "Bearer dev:owner", map[string]string{"Idempotency-Key": "job"}, map[string]any{"printerId": p.ID, "expectedMediaFingerprint": p.MediaFingerprint, "templateId": "qr-title", "templateVersion": 1, "templateOptions": map[string]any{"showReference": true}, "copies": 1})
	requireStatus(t, enqueue, 201)
	secret := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	claim := map[string]any{"printerId": p.ID, "attemptId": "attempt-for-printing", "sessionId": "session-for-printing", "claimToken": secret}
	for _, bad := range []string{"", "Bearer dev:owner", "Bearer invalid"} {
		requireStatus(t, performRequest(server, "POST", "/print-consumer/claims", bad, claim), 401)
	}
	claimed := performRequest(server, "POST", "/print-consumer/claims", token, claim)
	requireStatus(t, claimed, 200)
	data := labelResponseData(t, claimed.Body.Bytes())
	if data["attemptId"] != "attempt-for-printing" {
		t.Fatalf("claim %s", claimed.Body.String())
	}
	replay := performRequest(server, "POST", "/print-consumer/claims", token, claim)
	requireStatus(t, replay, 200)
	if labelResponseData(t, replay.Body.Bytes())["revision"] != data["revision"] {
		t.Fatal("claim retry changed revision")
	}
	path := "/print-consumer/claims/attempt-for-printing"
	proof := map[string]any{"sessionId": "session-for-printing", "claimToken": secret, "revision": data["revision"]}
	wrong := map[string]any{"sessionId": "different-session", "claimToken": secret, "revision": data["revision"]}
	requireStatus(t, performRequest(server, "POST", path+"/start", token, wrong), 403)
	headers := map[string]string{"X-Print-Session-ID": "session-for-printing", "X-Print-Claim-Token": secret, "X-Print-Revision": "2"}
	content := performRequestWithHeaders(server, "GET", path+"/content", token, headers, nil)
	requireStatus(t, content, 200)
	sum := sha256.Sum256(content.Body.Bytes())
	if hex.EncodeToString(sum[:]) != data["artifact"].(map[string]any)["sha256"] {
		t.Fatal("artifact digest mismatch")
	}
	for _, route := range []struct {
		method, path string
		body         any
	}{
		{"POST", path + "/start", proof}, {"POST", path + "/renewal", proof},
		{"POST", path + "/outcome", map[string]any{"sessionId": "session-for-printing", "claimToken": secret, "revision": 2, "outcome": map[string]any{"kind": "no_output", "completedCopies": 0, "retryable": false, "reason": "unknown"}}},
		{"GET", "/print-consumer/attempts/attempt-for-printing", nil}, {"GET", "/print-consumer/attempts", nil},
		{"POST", "/print-consumer/attempts/attempt-for-printing/reconciliation", map[string]any{"revision": 2, "outcome": map[string]any{"kind": "completed", "completedCopies": 1, "retryable": false, "reason": ""}}},
	} {
		requireStatus(t, performRequest(server, route.method, route.path, "Bearer invalid", route.body), 401)
	}
	requireStatus(t, performRequestWithHeaders(server, "GET", path+"/content", "Bearer invalid", headers, nil), 401)
	requireStatus(t, performRequest(server, "GET", "/print-consumer/attempts/not-owned-attempt", token, nil), 404)
	requireStatus(t, performRequest(server, "POST", path+"/renewal", token, wrong), 403)
	headers["X-Print-Session-ID"] = "different-session"
	requireStatus(t, performRequestWithHeaders(server, "GET", path+"/content", token, headers, nil), 403)
	headers["X-Print-Session-ID"] = "session-for-printing"
	discovery := performRequest(server, "GET", "/print-consumer/attempts?status=unsettled", token, nil)
	requireStatus(t, discovery, 200)
	if !strings.Contains(discovery.Body.String(), "attempt-for-printing") || strings.Contains(discovery.Body.String(), secret) || strings.Contains(discovery.Body.String(), "https://example.test") {
		t.Fatal("recovery discovery lost attempt or exposed label content")
	}
	renewed := performRequest(server, "POST", path+"/renewal", token, proof)
	requireStatus(t, renewed, 200)
	proof["revision"] = labelResponseData(t, renewed.Body.Bytes())["revision"]
	az.SetPrintingAvailable(false)
	failed := performRequest(server, "POST", path+"/start", token, proof)
	if failed.Code < 400 {
		t.Fatal("authorization outage allowed start")
	}
	az.SetPrintingAvailable(true)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/printer-reports", token, map[string]any{"printerId": p.ID, "state": "unavailable"}), 200)
	requireStatus(t, performRequest(server, "POST", path+"/start", token, proof), 409)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/printer-reports", token, map[string]any{"printerId": p.ID, "state": "ready"}), 200)
	clock.now = clock.now.Add(2 * time.Second)
	requireStatus(t, performRequest(server, "POST", path+"/start", token, proof), 409)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/heartbeat", token, map[string]any{"sessionId": "session-for-printing"}), 200)
	requireStatus(t, performRequest(server, "POST", path+"/start", token, proof), 409)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/printer-reports", token, map[string]any{"printerId": p.ID, "state": "ready"}), 200)
	started := performRequest(server, "POST", path+"/start", token, proof)
	requireStatus(t, started, 200)
	view := performRequest(server, "GET", "/print-consumer/attempts/attempt-for-printing", token, nil)
	requireStatus(t, view, 200)
	recovered := labelResponseData(t, view.Body.Bytes())
	if recovered["status"] != "printing" || recovered["sessionId"] != "session-for-printing" {
		t.Fatal(view.Body.String())
	}
	proof["revision"] = labelResponseData(t, started.Body.Bytes())["revision"]
	proof["outcome"] = map[string]any{"kind": "completed", "completedCopies": 1, "retryable": false, "reason": ""}
	done := performRequest(server, "POST", path+"/outcome", token, proof)
	requireStatus(t, done, http.StatusOK)
	if labelResponseData(t, done.Body.Bytes())["status"] != "completed" {
		t.Fatal(done.Body.String())
	}
	requireStatus(t, performRequest(server, "POST", path+"/outcome", token, proof), 200)
	requireStatus(t, performRequestWithHeaders(server, "GET", path+"/content", token, headers, nil), 409)
	// A started lease expiring never releases a printer for another label.
	next := performRequestWithHeaders(server, "POST", labelPrefix+"/assets/"+assetID+"/print-jobs", "Bearer dev:owner", map[string]string{"Idempotency-Key": "after-restart"}, map[string]any{"printerId": p.ID, "expectedMediaFingerprint": p.MediaFingerprint, "templateId": "qr-only", "templateVersion": 1, "templateOptions": map[string]any{"showReference": false}, "copies": 1})
	requireStatus(t, next, 201)
	claim["attemptId"] = "attempt-after-restart"
	claimed = performRequest(server, "POST", "/print-consumer/claims", token, claim)
	requireStatus(t, claimed, 200)
	path = "/print-consumer/claims/attempt-after-restart"
	proof = map[string]any{"sessionId": "session-for-printing", "claimToken": secret, "revision": labelResponseData(t, claimed.Body.Bytes())["revision"]}
	started = performRequest(server, "POST", path+"/start", token, proof)
	requireStatus(t, started, 200)
	clock.now = clock.now.Add(6 * time.Second)
	expired := performRequest(server, "GET", "/print-consumer/attempts/attempt-after-restart", token, nil)
	requireStatus(t, expired, 200)
	if labelResponseData(t, expired.Body.Bytes())["leaseValid"] != false {
		t.Fatal("expired lease reported valid")
	}
	proof["revision"] = labelResponseData(t, started.Body.Bytes())["revision"]
	proof["outcome"] = map[string]any{"kind": "completed", "completedCopies": 1, "retryable": false, "reason": ""}
	requireStatus(t, performRequest(server, "POST", path+"/outcome", token, proof), 409)
	claim["attemptId"] = "attempt-no-replay-now"
	blocked := performRequest(server, "POST", "/print-consumer/claims", token, claim)
	requireStatus(t, blocked, 200)
	if !strings.Contains(blocked.Body.String(), "\"data\":null") {
		t.Fatal("uncertain reservation released")
	}
	uncertain := performRequest(server, "GET", "/print-consumer/attempts/attempt-after-restart", token, nil)
	requireStatus(t, uncertain, 200)
	reconciliation := map[string]any{"revision": labelResponseData(t, uncertain.Body.Bytes())["revision"], "outcome": proof["outcome"]}
	result := performRequest(server, "POST", "/print-consumer/attempts/attempt-after-restart/reconciliation", token, reconciliation)
	requireStatus(t, result, 200)
	if labelResponseData(t, result.Body.Bytes())["status"] != "completed" {
		t.Fatal(result.Body.String())
	}
	requireStatus(t, performRequest(server, "POST", "/print-consumer/attempts/attempt-after-restart/reconciliation", token, reconciliation), 200)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/heartbeat", token, map[string]any{"sessionId": "session-for-printing"}), 200)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/printer-reports", token, map[string]any{"printerId": p.ID, "state": "ready"}), 200)
	// Losing the initiating editor's grant cancels unstarted output.
	editor := identity.Principal{ID: "editor"}
	if err = az.GrantInventoryEditor(ctx, editor, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	queued := performRequestWithHeaders(server, "POST", labelPrefix+"/assets/"+assetID+"/print-jobs", "Bearer dev:editor", map[string]string{"Idempotency-Key": "editor-job"}, map[string]any{"printerId": p.ID, "expectedMediaFingerprint": p.MediaFingerprint, "templateId": "qr-only", "templateVersion": 1, "templateOptions": map[string]any{"showReference": false}, "copies": 1})
	requireStatus(t, queued, 201)
	claim["attemptId"] = "attempt-editor-revoked"
	claimed = performRequest(server, "POST", "/print-consumer/claims", token, claim)
	requireStatus(t, claimed, 200)
	if err = az.RevokeInventoryEditor(ctx, editor, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	proof = map[string]any{"sessionId": "session-for-printing", "claimToken": secret, "revision": labelResponseData(t, claimed.Body.Bytes())["revision"]}
	requireStatus(t, performRequest(server, "POST", "/print-consumer/claims/attempt-editor-revoked/start", token, proof), 409)
	canceled := performRequest(server, "GET", labelPrefix+"/print-jobs/"+labelResponseData(t, queued.Body.Bytes())["id"].(string), "Bearer dev:owner", nil)
	requireStatus(t, canceled, 200)
	if labelResponseData(t, canceled.Body.Bytes())["status"] != "canceled" {
		t.Fatal("revoked initiator left dispatchable work")
	}

	for _, operation := range []string{
		"POST /print-consumer/claims", "POST /print-consumer/claims/{attemptId}/renewal", "POST /print-consumer/claims/{attemptId}/start", "POST /print-consumer/claims/{attemptId}/outcome", "GET /print-consumer/claims/{attemptId}/content", "GET /print-consumer/attempts/{attemptId}", "GET /print-consumer/attempts", "POST /print-consumer/attempts/{attemptId}/reconciliation",
	} {
		coverage.operation[operation] = struct{}{}
	}
}
