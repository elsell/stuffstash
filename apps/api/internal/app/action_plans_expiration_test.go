package app

import (
	"context"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestCreateActionPlanExpirationValidation(t *testing.T) {
	for _, raw := range []string{
		`{"title":"Bottle","expiration":{"date":"2028-02","precision":"month"}}`,
		`{"title":"Bottle","customAssetTypeId":"medicine","expiration":null}`,
		`{"title":"Bottle","customAssetTypeId":"medicine","expiration":{"date":"2028-02-30","precision":"day"}}`,
		`{"title":"Bottle","customAssetTypeId":"medicine","expiration":{"date":"2028-02","precision":"month","extra":true}}`,
	} {
		if _, err := parseActionPlanCreateArguments(ports.ActionPlanCommandRecord{Kind: actionplan.CommandKindCreateAsset, ArgumentsJSON: []byte(raw)}); err == nil {
			t.Fatalf("accepted invalid expiration: %s", raw)
		}
	}
}
func TestCreateActionPlanExecutionPreservesExpiration(t *testing.T) {
	for _, precision := range []string{"day", "month"} {
		t.Run(precision, func(t *testing.T) {
			date := "2028-02"
			if precision == "day" {
				date += "-29"
			}
			record := actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindCreateAsset, `{"title":"Bottle","kind":"item","customAssetTypeId":"medicine","expiration":{"date":"`+date+`","precision":"`+precision+`"}}`)
			repository := &fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{"plan-1": record}}
			assets := &fakeAssetRepository{}
			application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"asset-1", "undo-1", "audit-1"}})
			kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
			kind.ExpirationEnabled = true
			types := &fakeCustomAssetTypeRepository{items: []customfield.AssetType{kind}}
			application.assetService = assetapp.New(assetapp.Dependencies{Authorizer: application.authorizer, Tenants: application.tenants, Inventories: application.inventories, Assets: assets, CustomAssetTypes: types, AssetUnitOfWork: assets, Undoables: assets, IDs: application.ids, Clock: application.clock, Audit: application.audit})
			result, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", PlanID: "plan-1"})
			if err != nil {
				t.Fatal(err)
			}
			item := assets.items[asset.ID("asset-1")]
			if result.State != actionplan.StateExecuted || item.CustomAssetTypeID != "medicine" || item.Expiration.Value() != date || string(item.Expiration.Precision()) != precision {
				t.Fatalf("expiration lost: %+v", item)
			}
		})
	}
}

func TestDependentExpirationCreateRevalidatesTypeBeforeAtomicWrite(t *testing.T) {
	for _, mode := range []string{"enabled", "disabled", "archived", "foreign"} {
		t.Run(mode, func(t *testing.T) {
			record := actionPlanRecord("plan-1", actionplan.StateApproved)
			record.Commands = []ports.ActionPlanCommandRecord{
				{ID: "closet", Kind: actionplan.CommandKindCreateLocation, Summary: "Create closet", ArgumentsJSON: []byte(`{"title":"Closet"}`)},
				{ID: "bottle", Kind: actionplan.CommandKindCreateAsset, Summary: "Add bottle", ArgumentsJSON: []byte(`{"title":"Bottle","parentCommandId":"closet","customAssetTypeId":"medicine","expiration":{"date":"2028-02","precision":"month"}}`)},
			}
			repository := &fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{"plan-1": record}}
			assets := &fakeAssetRepository{}
			application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{})
			kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
			kind.ExpirationEnabled = mode != "disabled"
			if mode == "archived" {
				kind.LifecycleState = customfield.AssetTypeLifecycleArchived
			}
			if mode == "foreign" {
				kind.TenantID = "other-tenant"
			}
			types := &fakeCustomAssetTypeRepository{items: []customfield.AssetType{kind}}
			application.assetService = assetapp.New(assetapp.Dependencies{Authorizer: application.authorizer, Tenants: application.tenants, Inventories: application.inventories, Assets: assets, CustomAssetTypes: types, AssetUnitOfWork: assets, Undoables: assets, IDs: application.ids, Clock: application.clock, Audit: application.audit})
			_, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", PlanID: "plan-1"})
			if mode != "enabled" {
				if err == nil || len(assets.items) != 0 || len(assets.auditRecords) != 0 {
					t.Fatalf("invalid type partially saved: %v", err)
				}
				return
			}
			if err != nil || len(assets.items) != 2 {
				t.Fatalf("dependent create: %v", err)
			}
			var bottle asset.Asset
			for _, item := range assets.items {
				if item.Title.String() == "Bottle" {
					bottle = item
				}
			}
			if bottle.Expiration.Value() != "2028-02" || bottle.ParentAssetID == "" || assets.items[bottle.ParentAssetID].Kind != asset.KindLocation {
				t.Fatal("dependent item lost expiration or parent")
			}
			foundUndo := false
			for _, undo := range assets.undoables {
				if undo.AfterAsset.ID == bottle.ID {
					foundUndo = true
					if undo.AfterAsset.Expiration != bottle.Expiration {
						t.Fatal("undo snapshot lost expiration")
					}
				}
			}
			if !foundUndo {
				t.Fatal("missing undo snapshot")
			}
		})
	}
}
