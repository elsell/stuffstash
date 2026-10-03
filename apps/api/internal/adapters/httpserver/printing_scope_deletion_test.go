package httpserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"github.com/stuffstash/stuff-stash/internal/adapters/pairingcrypto"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
	"time"
)

func TestInventoryDeletionRevokesExistingPrintConsumerAndBlocksScopeReuse(t *testing.T) {
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

	if err := az.GrantInventoryViewer(ctx, identity.Principal{ID: "viewer"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	if err := az.GrantInventoryEditor(ctx, identity.Principal{ID: "editor"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	body := map[string]any{"printerId": string(p.ID), "expectedMediaFingerprint": p.MediaFingerprint, "templateId": "qr-title", "templateVersion": 1, "templateOptions": map[string]any{"showReference": true}, "copies": 1}
	queued := performRequestWithHeaders(server, "POST", labelPrefix+"/printers/"+string(p.ID)+"/test-jobs", "Bearer dev:owner", map[string]string{"Idempotency-Key": "delete-test"}, body)
	requireStatus(t, queued, 201)
	job := labelResponseData(t, queued.Body.Bytes())
	for _, c := range []struct {
		token string
		code  int
	}{{"", 401}, {"Bearer dev:viewer", 403}, {"Bearer dev:editor", 403}, {"Bearer dev:other", 403}} {
		requireStatus(t, performRequest(server, "DELETE", labelPrefix, c.token, nil), c.code)
	}
	requireStatus(t, performRequest(server, "DELETE", labelPrefix, "Bearer dev:owner", nil), 204)
	requireStatus(t, performRequest(server, "POST", "/print-consumer/heartbeat", token, map[string]any{"sessionId": "session-for-printing"}), 401)
	requireStatus(t, performRequest(server, "GET", labelPrefix+"/printers", "Bearer dev:owner", nil), 404)
	requireStatus(t, performRequest(server, "GET", labelPrefix+"/print-jobs/"+job["id"].(string), "Bearer dev:owner", nil), 404)
	requireStatus(t, performRequestWithHeaders(server, "POST", labelPrefix+"/printers", "Bearer dev:owner", map[string]string{"Idempotency-Key": "new-after-delete"}, map[string]any{"name": "Garage", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1}), 404)
	preserved, err := store.GetPrintJob(ctx, actor.Scope, printing.JobID(job["id"].(string)))
	if err != nil || preserved.Status != printing.JobCanceled {
		t.Fatal("deletion did not preserve cancellation history")
	}
	// Relationship cleanup is independent of the now-absent human scope.
	if err = application.PrintConnectors().DrainAuthorization(ctx, 10); err != nil {
		t.Fatal(err)
	}
	pending, err := store.PendingPrintConnectorScopes(ctx, 10)
	if err != nil || len(pending) != 0 {
		t.Fatal("deleted scope stranded authorization removal")
	}
}
