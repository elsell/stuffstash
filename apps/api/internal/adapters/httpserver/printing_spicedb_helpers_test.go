package httpserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/pairingcrypto"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	"github.com/stuffstash/stuff-stash/internal/adapters/spicedb"
	"github.com/stuffstash/stuff-stash/internal/app"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type spicePrintFixture struct {
	t                    *testing.T
	ctx                  context.Context
	application          app.App
	store                *memory.Store
	az                   spicedb.Authorizer
	clock                *labelTestClock
	server               *http.Server
	actor                printregistry.Actor
	printer              printing.Printer
	prefix, owner, token string
	connector            printing.Connector
	policy               printregistry.ConnectorPolicy
}

func newSpicePrintFixture(t *testing.T) *spicePrintFixture {
	t.Helper()
	endpoint := os.Getenv("STUFF_STASH_SPICEDB_INTEGRATION_ENDPOINT")
	if endpoint == "" {
		t.Skip("STUFF_STASH_SPICEDB_INTEGRATION_ENDPOINT is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	t.Cleanup(cancel)
	gateway, err := spicedb.NewGateway(endpoint, "", false, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = gateway.Close() })
	az := spicedb.NewAuthorizer(gateway)
	schemaPath := os.Getenv("STUFF_STASH_SPICEDB_INTEGRATION_SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = filepath.Join("..", "..", "..", "..", "..", "deploy", "spicedb", "schema.zed")
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = az.BootstrapSchema(ctx, string(schema)); err != nil {
		t.Fatal(err)
	}
	ids := idgen.NewULIDGenerator()
	tid, iid, owner := ids.NewID(), ids.NewID(), ids.NewID()
	clock := &labelTestClock{now: time.Now().UTC()}
	store := memory.NewStore()
	seedMemoryStore(t, ctx, store, az, seededState{tenants: []seedTenant{{id: tid, name: "Home", owner: owner}}, inventories: []seedInventory{{id: iid, tenantID: tid, name: "Tools", owner: owner}}})
	renderer, err := labelrenderer.New(labelrenderer.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	labels := printingapp.NewLabelService(printingapp.LabelDependencies{Repository: store, Renders: store, Assets: store, Inventories: store, Tenants: store, Authorizer: az, Audit: store, IDs: ids, Clock: clock, Renderer: renderer, Templates: renderer, BaseURL: "https://labels.example.test", RenderTTL: time.Hour, MaxRenderBytes: 1000000})
	if _, err = labels.BootstrapInstance(ctx); err != nil {
		t.Fatal(err)
	}
	application := app.New(app.Dependencies{Clock: clock, Labels: labels, Observer: &fakeObserver{}, Auth: auth.NewLocalDevAuthenticator(), Authorizer: az, Users: store, Tenants: store, TenantUnitOfWork: store, Inventories: store, InventoryUnitOfWork: store, InventoryAccess: store, InventoryAccessUnitOfWork: store, Assets: store, AssetUnitOfWork: store, AssetTags: store, AssetTagUnitOfWork: store, Checkouts: store, Undoables: store, CustomAssetTypes: store, CustomFields: store, Audit: store, Outbox: store})
	application = application.WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintSettings(store).WithPrintJobs(store, printingapp.JobConfig{MaxCopies: 20, MaxArtifactBytes: 1000000, ArtifactTTL: time.Hour, Lease: 5 * time.Second, ReadinessMaxAge: time.Hour})
	f := &spicePrintFixture{t: t, ctx: ctx, application: application, store: store, az: az, clock: clock, owner: "Bearer dev:" + owner, prefix: "/tenants/" + tid + "/inventories/" + iid, actor: printregistry.Actor{Principal: identity.Principal{ID: identity.PrincipalID(owner)}, Scope: printing.Scope{TenantID: tid, InventoryID: iid}}, policy: printregistry.ConnectorPolicy{PublicWebBaseURL: "https://example.test", PairingLifetime: time.Minute, CredentialLifetime: time.Hour, ActivationLifetime: time.Minute, AuthorizationTimeout: 100 * time.Millisecond, ReportMaxAge: time.Hour}}
	f.authorization(az)
	p, _, err := f.application.PrinterRegistry().Register(ctx, printregistry.RegisterPrinter{Actor: f.actor, Name: "Garage", RequestKey: ids.NewID(), AdapterID: "brother-ql800", PresetID: "brother-ql800-29x90", PresetVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	f.printer = p
	return f
}
func (f *spicePrintFixture) authorization(az ports.PrintingAuthorization) {
	f.application = f.application.WithPrintConnectors(f.store, az, pairingcrypto.Secrets{}, f.policy)
	f.server = NewServer(":0", f.application)
}
func (f *spicePrintFixture) pair() (printregistry.PairingStarted, ed25519.PrivateKey) {
	f.t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		f.t.Fatal(err)
	}
	pair, err := f.application.PrintConnectors().Begin(f.ctx, printregistry.BeginPairing{Name: "USB host", PublicKey: pub, Candidates: []printing.PairingCandidate{{ID: "usb", Name: "Brother", AdapterID: "brother-ql800", DeviceID: "test-physical-device"}}})
	if err != nil {
		f.t.Fatal(err)
	}
	_, err = f.application.PrintConnectors().Approve(f.ctx, printregistry.ApprovePairing{Actor: f.actor, PairingID: pair.Pairing.ID, UserCode: pair.UserCode, Bindings: []printregistry.PairingBinding{{CandidateID: "usb", PrinterID: f.printer.ID}}})
	if err != nil {
		f.t.Fatal(err)
	}
	return pair, key
}
func (f *spicePrintFixture) activate(pair printregistry.PairingStarted, key ed25519.PrivateKey) {
	f.t.Helper()
	credential, err := f.application.PrintConnectors().Exchange(f.ctx, pair.Pairing.ID, pair.PollToken, ed25519.Sign(key, printing.PairingProofMessage(pair.Pairing.ID, pair.PollToken)))
	if err != nil {
		f.t.Fatal(err)
	}
	f.connector = credential.Connector
	f.token = "Bearer " + credential.Credential
	requireStatus(f.t, performRequest(f.server, "POST", "/print-consumer/heartbeat", f.token, map[string]any{"sessionId": "lifecycle-session"}), 200)
	requireStatus(f.t, performRequest(f.server, "POST", "/print-consumer/printer-reports", f.token, map[string]any{"printerId": f.printer.ID, "state": "ready"}), 200)
}
func (f *spicePrintFixture) queuedClaim(token string) (string, string, map[string]any) {
	f.t.Helper()
	id := idgen.NewULIDGenerator().NewID()
	asset := performRequest(f.server, "POST", f.prefix+"/assets", f.owner, map[string]any{"title": "Tool", "kind": "item"})
	requireStatus(f.t, asset, 201)
	job := performRequestWithHeaders(f.server, "POST", f.prefix+"/assets/"+decodeAsset(f.t, asset).Data.ID+"/print-jobs", token, map[string]string{"Idempotency-Key": id}, map[string]any{"printerId": f.printer.ID, "expectedMediaFingerprint": f.printer.MediaFingerprint, "templateId": "qr-title", "templateVersion": 1, "templateOptions": map[string]any{"showReference": true}, "copies": 1})
	requireStatus(f.t, job, 201)
	secret := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	claim := performRequest(f.server, "POST", "/print-consumer/claims", f.token, map[string]any{"printerId": f.printer.ID, "attemptId": id, "sessionId": "lifecycle-session", "claimToken": secret})
	requireStatus(f.t, claim, 200)
	data := labelResponseData(f.t, claim.Body.Bytes())
	if data == nil {
		f.t.Fatal("no claim")
	}
	return labelResponseData(f.t, job.Body.Bytes())["id"].(string), id, map[string]any{"sessionId": "lifecycle-session", "claimToken": secret, "revision": data["revision"]}
}
