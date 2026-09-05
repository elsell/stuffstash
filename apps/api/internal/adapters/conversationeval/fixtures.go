package conversationeval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type isolatedRuntime struct {
	application app.App
	store       *memory.Store
	tenantID    tenant.ID
	inventoryID inventory.InventoryID
	runtimeIDs  map[string]string
}

func (r isolatedRuntime) assertUnchanged(ctx context.Context, before []byte) error {
	after, err := r.snapshot(ctx)
	if err != nil {
		return err
	}
	if !bytes.Equal(before, after) {
		return ErrFixtureMutation
	}
	return nil
}

func (e *Executor) prepare(ctx context.Context, input ports.ConversationEvaluationInput, calls *atomic.Int64) (isolatedRuntime, error) {
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	tenantID := tenant.ID(input.Revision.Snapshot().TenantID)
	inventoryID, ok := inventory.NewID(e.deps.IDs.NewID())
	if !ok {
		return isolatedRuntime{}, ErrInvalidExecution
	}
	tenantName, ok := tenant.NewName("Evaluation")
	if !ok {
		return isolatedRuntime{}, ErrInvalidExecution
	}
	inventoryName, ok := inventory.NewName("Evaluation fixtures")
	if !ok {
		return isolatedRuntime{}, ErrInvalidExecution
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenantID, Name: tenantName}); err != nil {
		return isolatedRuntime{}, err
	}
	if err := store.SaveInventory(ctx, inventory.Inventory{ID: inventoryID, TenantID: inventory.TenantID(tenantID), Name: inventoryName}); err != nil {
		return isolatedRuntime{}, err
	}
	if err := authorizer.GrantTenantOwner(ctx, input.Principal, tenantID); err != nil {
		return isolatedRuntime{}, err
	}
	if err := authorizer.GrantInventoryOwner(ctx, input.Principal, tenantID, inventoryID); err != nil {
		return isolatedRuntime{}, err
	}
	application := app.New(app.Dependencies{Authorizer: authorizer, Tenants: store, Inventories: store, Assets: store, AssetUnitOfWork: store, AssetTags: store, AssetTagUnitOfWork: store, Undoables: store, Search: store, Checkouts: store, Audit: store, ActionPlans: store, RealtimeSessions: store, IDs: e.deps.IDs, Clock: e.deps.Clock, Observer: e.deps.Observer, ConversationWorkflows: pinnedWorkflow{revision: input.Revision}, ConversationWorkflowLimits: input.Limits, RealtimeVoiceProviderResolver: textProviders{providers: input.Providers, explicit: input.WorkflowProviders, transcript: input.Case.Settings().Utterance, calls: calls}})
	fixtureIDs := map[string]string{}
	tags := map[string]string{}
	pending := input.Case.Settings().Assets
	for len(pending) > 0 {
		remaining := pending[:0]
		for _, fixture := range pending {
			parentID := fixtureIDs[fixture.ParentID]
			if fixture.ParentID != "" && parentID == "" {
				remaining = append(remaining, fixture)
				continue
			}
			tagIDs := []string{}
			for _, name := range fixture.TagNames {
				id, exists := tags[name]
				if !exists {
					tag, err := application.CreateAssetTag(ctx, app.CreateAssetTagInput{Principal: input.Principal, TenantID: tenantID, InventoryID: inventoryID, Source: audit.SourceSystem, Key: fmt.Sprintf("fixture-tag-%d", len(tags)+1), DisplayName: name, Color: "#2f80ed"})
					if err != nil {
						return isolatedRuntime{}, err
					}
					id = tag.ID.String()
					tags[name] = id
				}
				tagIDs = append(tagIDs, id)
			}
			created, err := application.CreateAssetWithOperation(ctx, app.CreateAssetInput{Principal: input.Principal, TenantID: tenantID, InventoryID: inventoryID, Source: audit.SourceSystem, Title: fixture.Title, Kind: string(fixture.Kind), Description: fixture.Description, ParentAssetID: parentID, TagIDs: tagIDs})
			if err != nil {
				return isolatedRuntime{}, err
			}
			fixtureIDs[fixture.ID] = created.Asset.ID.String()
		}
		if len(remaining) == len(pending) {
			return isolatedRuntime{}, ErrInvalidExecution
		}
		pending = remaining
	}
	runtimeIDs := map[string]string{}
	for fixtureID, runtimeID := range fixtureIDs {
		runtimeIDs[runtimeID] = fixtureID
	}
	return isolatedRuntime{application: application, store: store, tenantID: tenantID, inventoryID: inventoryID, runtimeIDs: runtimeIDs}, nil
}
func (r isolatedRuntime) snapshot(ctx context.Context) ([]byte, error) {
	assets, err := r.store.ListAssetsByInventory(ctx, r.tenantID, r.inventoryID, ports.AssetListPageRequest{Limit: domain.MaxEvaluationFixtureAssets + 1, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		return nil, err
	}
	ids := make([]asset.ID, len(assets))
	for index, item := range assets {
		ids[index] = item.ID
	}
	tags, err := r.store.AssetTagsByAssets(ctx, r.tenantID, r.inventoryID, ids)
	if err != nil {
		return nil, err
	}
	checkouts, err := r.store.CurrentAssetCheckouts(ctx, r.tenantID, r.inventoryID, ids)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Assets    []asset.Asset
		Tags      map[asset.ID][]assettag.Tag
		Checkouts map[asset.ID]asset.Checkout
	}{assets, tags, checkouts})
}
