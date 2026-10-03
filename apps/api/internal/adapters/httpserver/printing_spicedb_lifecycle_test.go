package httpserver

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/spicedb"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestSpiceDBIntegrationPrintPermissionLossBeforeStart(t *testing.T) {
	f := newSpicePrintFixture(t)
	pair, key := f.pair()
	f.activate(pair, key)
	editor := identity.Principal{ID: identity.PrincipalID(idgen.NewULIDGenerator().NewID())}
	tid, iid := tenant.ID(f.actor.Scope.TenantID), inventory.InventoryID(f.actor.Scope.InventoryID)
	if err := f.az.GrantInventoryEditor(f.ctx, editor, tid, iid); err != nil {
		t.Fatal(err)
	}
	job, attempt, proof := f.queuedClaim("Bearer dev:" + string(editor.ID))
	if err := f.az.RevokeInventoryEditor(f.ctx, editor, tid, iid); err != nil {
		t.Fatal(err)
	}
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/claims/"+attempt+"/start", f.token, proof), 409)
	current := performRequest(f.server, "GET", f.prefix+"/print-jobs/"+job, f.owner, nil)
	requireStatus(t, current, 200)
	if labelResponseData(t, current.Body.Bytes())["status"] != "canceled" {
		t.Fatal("permission loss did not cancel unstarted output", current.Body.String())
	}
}

func TestSpiceDBIntegrationPrintCredentialRotationPreservesPrivileges(t *testing.T) {
	f := newSpicePrintFixture(t)
	pair, key := f.pair()
	f.activate(pair, key)
	original := f.connector
	other, _, err := f.application.PrinterRegistry().Register(f.ctx, printregistry.RegisterPrinter{Actor: f.actor, Name: "Unassigned", RequestKey: "unassigned", AdapterID: "brother-ql800", PresetID: "brother-ql800-29x90", PresetVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	pub, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rotation, err := f.application.PrintConnectors().Begin(f.ctx, printregistry.BeginPairing{Rotation: true, Name: "Replacement", PublicKey: pub})
	if err != nil {
		t.Fatal(err)
	}
	requireStatus(t, performRequest(f.server, "POST", f.prefix+"/print-connectors/"+string(original.ID)+"/credential-rotation", f.owner, map[string]any{"generation": original.Generation, "pairingId": rotation.Pairing.ID, "userCode": rotation.UserCode}), 200)
	replacement, err := f.application.PrintConnectors().Exchange(f.ctx, rotation.Pairing.ID, rotation.PollToken, ed25519.Sign(private, printing.PairingProofMessage(rotation.Pairing.ID, rotation.PollToken)))
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Connector.ID != original.ID || replacement.Connector.ServiceAccountID != original.ServiceAccountID {
		t.Fatal("rotation changed principal")
	}
	next := "Bearer " + replacement.Credential
	requireStatus(t, performRequest(f.server, "GET", "/print-consumer/printers", next, nil), 403)
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/heartbeat", next, map[string]any{"sessionId": "replacement-session"}), 200)
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/heartbeat", f.token, map[string]any{"sessionId": "lifecycle-session"}), 401)
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/printer-reports", next, map[string]any{"printerId": f.printer.ID, "state": "ready"}), 200)
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/printer-reports", next, map[string]any{"printerId": other.ID, "state": "ready"}), 403)
	requireStatus(t, performRequest(f.server, "GET", f.prefix+"/assets", next, nil), 401)
	for _, permission := range []ports.PrinterPermission{ports.PrinterPermissionViewConsumer, ports.PrinterPermissionReport, ports.PrinterPermissionConsume} {
		if err := f.az.CheckPrinter(f.ctx, original.ServiceAccountID, f.printer.ID, permission); err != nil {
			t.Fatal("rotation lost assigned privilege", err)
		}
		if err := f.az.CheckPrinter(f.ctx, original.ServiceAccountID, other.ID, permission); !errors.Is(err, ports.ErrForbidden) {
			t.Fatal("rotation expanded privileges", err)
		}
	}
}

func TestSpiceDBIntegrationPrintPendingDesiredStateAndReorderedNotifications(t *testing.T) {
	f := newSpicePrintFixture(t)
	// A genuinely unavailable gRPC endpoint: authorization checks themselves still
	// use real SpiceDB; only relationship synchronization is temporarily offline.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err = listener.Close(); err != nil {
		t.Fatal(err)
	}
	gateway, err := spicedb.NewGateway(address, "", false, "")
	if err != nil {
		t.Fatal(err)
	}
	defer gateway.Close()
	f.authorization(spicedb.NewAuthorizer(gateway))
	pair, key := f.pair()
	signature := base64.StdEncoding.EncodeToString(ed25519.Sign(key, printing.PairingProofMessage(pair.Pairing.ID, pair.PollToken)))
	requireStatus(t, performRequestWithHeaders(f.server, "POST", "/print-connector-pairings/"+string(pair.Pairing.ID)+"/credential", "", map[string]string{"X-Pairing-Token": pair.PollToken}, map[string]any{"signature": signature}), 409)
	pending, err := f.application.PrintConnectors().Poll(f.ctx, pair.Pairing.ID, pair.PollToken)
	if err != nil {
		t.Fatal(err)
	}
	registration, err := f.store.GetPrintConnector(f.ctx, pending.Scope, pending.ConnectorID)
	if err != nil {
		t.Fatal(err)
	}
	if registration.Connector.Generation == registration.Connector.SyncedGeneration {
		t.Fatal("failed synchronization recorded as applied")
	}
	if err := f.az.CheckPrinter(f.ctx, registration.Connector.ServiceAccountID, f.printer.ID, ports.PrinterPermissionConsume); !errors.Is(err, ports.ErrForbidden) {
		t.Fatal("pending grant already permits consumption", err)
	}
	f.authorization(f.az)
	if err = f.application.PrintConnectors().DrainAuthorization(f.ctx, 10); err != nil {
		t.Fatal(err)
	}
	f.activate(pair, key)
	// Persist revoke/grant/revoke while delivery is unavailable. The historical
	// notifications contain an ID, never an old permission snapshot.
	f.authorization(spicedb.NewAuthorizer(gateway))
	generation := f.connector.Generation
	for _, selected := range [][]printing.PrinterID{{}, {f.printer.ID}, {}} {
		result := performRequest(f.server, "PATCH", f.prefix+"/print-connectors/"+string(f.connector.ID), f.owner, map[string]any{"generation": generation, "printerIds": selected})
		requireStatus(t, result, 200)
		generation++
	}
	if err := f.az.CheckPrinter(f.ctx, f.connector.ServiceAccountID, f.printer.ID, ports.PrinterPermissionConsume); err != nil {
		t.Fatal("test did not preserve stale real grant", err)
	}
	f.authorization(f.az)
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/printer-reports", f.token, map[string]any{"printerId": f.printer.ID, "state": "ready"}), 401)
	for range 3 {
		if err := f.application.PrintConnectors().Reconcile(f.ctx, f.actor.Scope, f.connector.ID); err != nil {
			t.Fatal(err)
		}
		if err := f.az.CheckPrinter(f.ctx, f.connector.ServiceAccountID, f.printer.ID, ports.PrinterPermissionConsume); !errors.Is(err, ports.ErrForbidden) {
			t.Fatal("late notification resurrected revoked grant", err)
		}
	}
	requireStatus(t, performRequest(f.server, "POST", "/print-consumer/printer-reports", f.token, map[string]any{"printerId": f.printer.ID, "state": "ready"}), 403)
}

func TestSpiceDBIntegrationRetiredPrinterReportsAndReconcilesStartedOutput(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		name := "outcome"
		if recovery {
			name = "reconciliation"
		}
		t.Run(name, func(t *testing.T) {
			f := newSpicePrintFixture(t)
			pair, key := f.pair()
			f.activate(pair, key)
			_, attempt, proof := f.queuedClaim(f.owner)
			started := performRequest(f.server, "POST", "/print-consumer/claims/"+attempt+"/start", f.token, proof)
			requireStatus(t, started, 200)
			proof["revision"] = labelResponseData(t, started.Body.Bytes())["revision"]
			requireStatus(t, performRequest(f.server, "PATCH", f.prefix+"/printers/"+string(f.printer.ID), f.owner, map[string]any{"revision": f.printer.Revision, "retired": true}), 200)
			outcome := map[string]any{"kind": "completed", "completedCopies": 1, "retryable": false, "reason": ""}
			path := "/print-consumer/claims/" + attempt + "/outcome"
			proof["outcome"] = outcome
			if recovery {
				f.clock.now = f.clock.now.Add(6 * time.Second)
				if _, err := f.application.PrintJobs().Maintain(f.ctx, ""); err != nil {
					t.Fatal(err)
				}
				current := performRequest(f.server, "GET", "/print-consumer/attempts/"+attempt, f.token, nil)
				requireStatus(t, current, 200)
				data := labelResponseData(t, current.Body.Bytes())
				if data["status"] != "uncertain" {
					t.Fatal(current.Body.String())
				}
				path = "/print-consumer/attempts/" + attempt + "/reconciliation"
				proof = map[string]any{"revision": data["revision"], "outcome": outcome}
			}
			result := performRequest(f.server, "POST", path, f.token, proof)
			requireStatus(t, result, 200)
			if labelResponseData(t, result.Body.Bytes())["status"] != "completed" {
				t.Fatal(result.Body.String())
			}
			requireStatus(t, performRequest(f.server, "POST", path, f.token, proof), 200)
			claim := performRequest(f.server, "POST", "/print-consumer/claims", f.token, map[string]any{"printerId": f.printer.ID, "attemptId": idgen.NewULIDGenerator().NewID(), "sessionId": "lifecycle-session", "claimToken": base64.RawURLEncoding.EncodeToString(make([]byte, 32))})
			requireStatus(t, claim, 409)
		})
	}
}
