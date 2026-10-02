package gormstore

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type expiringArchiveClock struct {
	now   time.Time
	calls int
}

func (c *expiringArchiveClock) Now() time.Time {
	c.calls++
	if c.calls > 1 {
		return c.now.Add(2 * time.Minute)
	}
	return c.now
}
func TestArchivePublicationRollsBackExpiredLeaseAndReservedBlob(t *testing.T) {
	for _, reason := range []string{"lease expires", "retired key"} {
		t.Run(reason, func(t *testing.T) {
			ctx := context.Background()
			s := newTestStore(t, ctx)
			saveTenant(t, ctx, s, "tenant", "Home")
			input, clock := archivePublicationFixture(t, s, "tenant", "restored", "restore")
			var publicationTime ports.Clock = clock
			if reason == "lease expires" {
				publicationTime = &expiringArchiveClock{now: clock.now}
			} else {
				if err := reserveMediaBlobKey(s.db, input.Plan.Document.Assets[0].Attachments[0].StorageKey); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := NewArchiveRestorePublisher(s, publicationTime, 100).PublishArchiveRestore(ctx, input); err == nil {
				t.Fatal("publication should fail")
			}
			if _, found, err := s.InventoryByID(ctx, "tenant", "restored"); err != nil || found {
				t.Fatal("failed publication left inventory")
			}
		})
	}
}
func archivePublicationFixture(t *testing.T, s Store, tid, iid, id string) (ports.ArchiveRestorePublication, *publicationClock) {
	t.Helper()
	clock := &publicationClock{now: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)}
	job := createRestoreClaimFor(t, s, clock.now, tid, iid, id)
	doc := ports.InventoryExportDocument{SchemaVersion: 2, TenantID: tid, InventoryID: iid, InventoryName: "Restored", ExportedAt: clock.now,
		CustomAssetTypes:       []customfield.AssetType{{ID: "pub-type", TenantID: customfield.TenantID(tid), InventoryID: customfield.InventoryID(iid), Scope: customfield.ScopeInventory, Key: "medicine", DisplayName: "Medicine", LifecycleState: customfield.AssetTypeLifecycleArchived}},
		CustomFieldDefinitions: []customfield.Definition{{ID: "pub-field", TenantID: customfield.TenantID(tid), InventoryID: customfield.InventoryID(iid), Scope: customfield.ScopeInventory, Key: "strength", DisplayName: "Strength", Type: customfield.FieldTypeNumber, Applicability: customfield.ApplicabilityCustomAssetTypes, CustomAssetTypeIDs: []customfield.AssetTypeID{"pub-type"}, LifecycleState: customfield.DefinitionLifecycleArchived}},
		Assets:                 []ports.InventoryExportAsset{{ID: "pub-asset", Title: "Bottle", Kind: "item", LifecycleState: "archived", CustomAssetTypeID: "pub-type", CustomFields: map[string]any{"strength": float64(5)}, CreatedAt: clock.now, UpdatedAt: clock.now, Attachments: []media.Attachment{{ID: "pub-photo", TenantID: media.TenantID(tid), InventoryID: media.InventoryID(iid), AssetID: "pub-asset", StorageKey: media.StorageKey(tid + "/" + iid + "/pub-asset/pub-photo"), FileName: "bottle.jpg", ContentType: media.ContentTypeJPEG, SHA256: media.SHA256(strings.Repeat("a", 64)), SizeBytes: 10, CreatedAt: clock.now, LifecycleState: media.LifecycleStateArchived}}}}}
	plan := ports.ArchiveRestorePlan{Document: doc}
	records, err := dataportability.BuildArchiveRestoreAudits(plan, job, &archiveAuditIDs{}, clock)
	if err != nil {
		t.Fatal(err)
	}
	return ports.ArchiveRestorePublication{Job: job, Plan: plan, OwnerGrantEventID: "archive-pub-grant", AuditRecords: records}, clock
}
