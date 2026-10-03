package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const labelTenant = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
const labelInventory = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
const labelOtherInventory = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
const labelPrefix = "/tenants/" + labelTenant + "/inventories/" + labelInventory

func labelTestServer(t *testing.T, clocks ...ports.Clock) (*http.Server, *memory.Store, *memory.Authorizer) {
	application, store, az := labelTestApplication(t, clocks...)
	return NewServer(":0", application), store, az
}
func labelTestApplication(t *testing.T, clocks ...ports.Clock) (app.App, *memory.Store, *memory.Authorizer) {
	t.Helper()
	ctx := context.Background()
	store := memory.NewStore()
	az := memory.NewAuthorizer()
	seedMemoryStore(t, ctx, store, az, seededState{tenants: []seedTenant{{id: labelTenant, name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: labelInventory, tenantID: labelTenant, name: "Tools", owner: "owner"}, {id: labelOtherInventory, tenantID: labelTenant, name: "Elsewhere", owner: "other"}}})
	clock := ports.Clock(ports.SystemClock{})
	if len(clocks) > 0 {
		clock = clocks[0]
	}
	renderer, _ := labelrenderer.New(labelrenderer.DefaultLimits())
	svc := printingapp.NewLabelService(printingapp.LabelDependencies{Repository: store, Renders: store, Assets: store, Inventories: store, Tenants: store, Authorizer: az, Audit: store, IDs: idgen.NewULIDGenerator(), Clock: clock, Renderer: renderer, Templates: renderer, BaseURL: "https://example.test/stash", RenderTTL: time.Hour, MaxRenderBytes: 1000000})
	if _, err := svc.BootstrapInstance(ctx); err != nil {
		t.Fatal(err)
	}
	application := app.New(app.Dependencies{Clock: clock, Labels: svc, Observer: &fakeObserver{}, Auth: auth.NewLocalDevAuthenticator(), Authorizer: az, Users: store, Tenants: store, TenantUnitOfWork: store, Inventories: store, InventoryUnitOfWork: store, InventoryAccess: store, InventoryAccessUnitOfWork: store, Assets: store, AssetUnitOfWork: store, AssetTags: store, AssetTagUnitOfWork: store, Checkouts: store, Undoables: store, CustomAssetTypes: store, CustomFields: store, Audit: store, Outbox: store})
	return application, store, az
}
func labelResponseData(t *testing.T, resBody []byte) map[string]any {
	t.Helper()
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(resBody, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data
}

type labelTestClock struct{ now time.Time }

func (c *labelTestClock) Now() time.Time { return c.now }
