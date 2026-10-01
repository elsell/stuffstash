package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCustomizationPlanCommitRollbackAndReplay(t *testing.T) {
	for _, field := range []bool{false, true} {
		for _, duplicateAudit := range []bool{false, true} {
			name := "type"
			if field {
				name = "field"
			}
			if duplicateAudit {
				name += "/rollback"
			} else {
				name += "/commit"
			}
			t.Run(name, func(t *testing.T) {
				ctx := context.Background()
				store := newTestStore(t, ctx)
				saveTenant(t, ctx, store, "tenant-home", "Home")
				saveInventory(t, ctx, store, "inventory-home", "tenant-home", "Home")
				plan := gormActionPlanRecord("plan-1", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
				plan.Commands[0].Kind = actionplan.CommandKindCreateCustomAssetType
				if field {
					plan.Commands[0].Kind = actionplan.CommandKindCreateCustomFieldDefinition
				}
				if err := store.SaveActionPlan(ctx, plan); err != nil {
					t.Fatal(err)
				}
				approveActionPlanForGormTest(t, ctx, store, plan)
				record := auditRecord(t, "schema-audit", plan.TenantID, plan.InventoryID, audit.ActionCustomAssetTypeCreated, audit.TargetCustomAssetType)
				if field {
					record.Action = audit.ActionCustomFieldDefinitionCreated
					record.TargetType = audit.TargetCustomFieldDefinition
				}
				if duplicateAudit {
					if err := store.SaveAuditRecord(ctx, record); err != nil {
						t.Fatal(err)
					}
				}
				transition := ports.ActionPlanStateTransition{PrincipalID: plan.PrincipalID, From: actionplan.StateApproved, To: actionplan.StateExecuted, At: plan.CreatedAt.Add(2 * time.Second)}
				execute := func() error {
					if field {
						definition, ok := customfield.NewDefinition("field-1", "tenant-home", "inventory-home", customfield.ScopeInventory, "warranty", "Warranty", customfield.FieldTypeDate, nil, customfield.ApplicabilityAllAssets, nil)
						if !ok {
							t.Fatal("invalid fixture")
						}
						_, _, err := store.ExecuteCreateCustomFieldActionPlan(ctx, plan.TenantID, plan.InventoryID, plan.ID, transition, definition, record)
						return err
					}
					kind, ok := customfield.NewAssetType("type-1", "tenant-home", "inventory-home", customfield.ScopeInventory, "warranty", "Warranty", "")
					if !ok {
						t.Fatal("invalid fixture")
					}
					_, _, err := store.ExecuteCreateCustomAssetTypeActionPlan(ctx, plan.TenantID, plan.InventoryID, plan.ID, transition, kind, record)
					return err
				}
				err := execute()
				if (err != nil) != duplicateAudit {
					t.Fatalf("wrong transaction result: %v", err)
				}
				saved, found, err := store.ActionPlanByID(ctx, plan.TenantID, plan.InventoryID, plan.ID)
				if err != nil || !found {
					t.Fatal(err)
				}
				expected := actionplan.StateExecuted
				if duplicateAudit {
					expected = actionplan.StateApproved
				}
				if saved.State != expected {
					t.Fatalf("partial plan transition: %s", saved.State)
				}
				exists := false
				if field {
					_, exists, err = store.CustomFieldDefinitionByID(ctx, plan.TenantID, plan.InventoryID, "field-1")
				} else {
					_, exists, err = store.CustomAssetTypeByID(ctx, plan.TenantID, plan.InventoryID, "type-1")
				}
				if err != nil || exists == duplicateAudit {
					t.Fatal("schema/audit/plan were not atomic")
				}
				if !duplicateAudit && execute() == nil {
					t.Fatal("replayed approval succeeded")
				}
			})
		}
	}
}
