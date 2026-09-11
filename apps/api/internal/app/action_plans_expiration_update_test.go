package app

import (
	"context"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestExpirationCorrectionRequiresExplicitValidDateOrClear(t *testing.T) {
	for _, raw := range []string{`{"assetId":"bottle"}`, `{"assetId":"bottle","expiration":{"date":"2028-02-30","precision":"day"}}`, `{"assetId":"bottle","expiration":{"date":"2028-02","precision":"month","unknown":true}}`, `{"assetId":"bottle","expiration":null,"title":"Other"}`, `{"expiration":null}`} {
		if err := validateExecutableActionPlanArguments(actionplan.CommandKindUpdateAsset, []byte(raw)); err == nil {
			t.Fatalf("invalid correction accepted: %s", raw)
		}
	}
	for _, raw := range []string{`{"assetId":"bottle","expiration":null}`, `{"assetId":"bottle","expiration":{"date":"2028-02","precision":"month"}}`} {
		if err := validateExecutableActionPlanArguments(actionplan.CommandKindUpdateAsset, []byte(raw)); err != nil {
			t.Fatalf("valid correction rejected: %v", err)
		}
	}
}
func TestApprovedExpirationCorrectionPreservesOtherFieldsAndUndo(t *testing.T) {
	for _, mode := range []string{"day", "month", "clear", "disabled", "unapproved", "foreign-tenant", "sibling-inventory"} {
		t.Run(mode, func(t *testing.T) {
			date := "2028-02"
			precision := "month"
			if mode == "day" {
				date += "-29"
				precision = "day"
			}
			value := `{"date":"` + date + `","precision":"` + precision + `"}`
			if mode == "clear" {
				value = "null"
				date = ""
			}
			state := actionplan.StateApproved
			if mode == "unapproved" {
				state = actionplan.StateProposed
			}
			record := actionPlanRecordWithCommand("plan-1", state, actionplan.CommandKindUpdateAsset, `{"assetId":"bottle","expiration":`+value+`}`)
			repository := &fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{"plan-1": record}}
			item := assetItem("bottle", "tenant-home", "inventory-home", asset.KindItem, "")
			item.CustomAssetTypeID = "medicine"
			item.Expiration, _ = expirationdate.ParseDate("2027-01", expirationdate.Month)
			if mode == "foreign-tenant" {
				item.TenantID = "other-tenant"
			}
			if mode == "sibling-inventory" {
				item.InventoryID = "other-inventory"
			}
			assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}}
			application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{})
			kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
			kind.ExpirationEnabled = mode != "disabled" && mode != "clear"
			types := &fakeCustomAssetTypeRepository{items: []customfield.AssetType{kind}}
			application.assetService = assetapp.New(assetapp.Dependencies{Authorizer: application.authorizer, Tenants: application.tenants, Inventories: application.inventories, Assets: assets, CustomAssetTypes: types, AssetUnitOfWork: assets, Undoables: assets, IDs: application.ids, Clock: application.clock, Audit: application.audit})
			result, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", PlanID: "plan-1"})
			saved := assets.items[item.ID]
			if mode == "disabled" || mode == "unapproved" || mode == "foreign-tenant" || mode == "sibling-inventory" {
				if err == nil || saved.Expiration != item.Expiration {
					t.Fatalf("invalid correction changed asset: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.State != actionplan.StateExecuted || saved.Expiration.Value() != date || saved.Title != item.Title || saved.ParentAssetID != item.ParentAssetID || saved.CustomAssetTypeID != item.CustomAssetTypeID {
				t.Fatalf("incorrect correction: %+v", saved)
			}
			found := false
			for _, undo := range assets.undoables {
				if undo.AfterAsset.ID == item.ID {
					found = true
					if undo.BeforeAsset.Expiration != item.Expiration || undo.AfterAsset.Expiration != saved.Expiration {
						t.Fatal("undo lost date")
					}
				}
			}
			if !found {
				t.Fatal("missing undo")
			}
			auditCount := len(assets.auditRecords)
			undoCount := len(assets.undoables)
			if _, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", PlanID: "plan-1"}); err == nil || len(assets.auditRecords) != auditCount || len(assets.undoables) != undoCount {
				t.Fatal("replay was not rejected without new writes")
			}
		})
	}
}
