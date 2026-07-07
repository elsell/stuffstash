package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCreateImportJobPreviewReportsAssetAndAttachmentSourceLinkDuplicates(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	existing := assetItem("asset-existing", "tenant-one", "inventory-one", asset.KindItem, "")
	customFields, ok := asset.NewCustomFields(map[string]any{"homebox-source-id": "source:drill"})
	if !ok {
		t.Fatalf("invalid custom fields")
	}
	existing.CustomFields = customFields
	if err := store.CreateAsset(ctx, existing, audit.Record{
		ID:          audit.ID("audit-existing-asset"),
		TenantID:    audit.TenantID("tenant-one"),
		InventoryID: audit.InventoryID("inventory-one"),
		Action:      audit.ActionAssetCreated,
		TargetType:  audit.TargetAsset,
		TargetID:    existing.ID.String(),
		OccurredAt:  now.Add(-time.Hour),
	}, nil); err != nil {
		t.Fatalf("seed existing asset: %v", err)
	}
	sourceIdentity := importSourceIdentity{sourceType: importplan.SourceLegacyHomebox, sourceInstanceKey: "https://homebox.example.test"}
	for _, input := range []importImportedResourceInput{
		{
			TenantID:         tenant.ID("tenant-one"),
			InventoryID:      inventory.InventoryID("inventory-one"),
			JobID:            importjob.ID("previous-job"),
			SourceIdentity:   sourceIdentity,
			SourceEntityType: ports.ImportSourceEntityAsset,
			SourceEntityID:   "source:drill",
			ResourceType:     ports.ImportResourceAsset,
			ResourceID:       "asset-existing",
			CreatedAt:        now.Add(-time.Minute),
		},
		{
			TenantID:         tenant.ID("tenant-one"),
			InventoryID:      inventory.InventoryID("inventory-one"),
			JobID:            importjob.ID("previous-job"),
			SourceIdentity:   sourceIdentity,
			SourceEntityType: ports.ImportSourceEntityAttachment,
			SourceEntityID:   "attachment:source:drill",
			ResourceType:     ports.ImportResourceAttachment,
			ResourceID:       "attachment-existing",
			ResourceOwnerID:  "asset-existing",
			CreatedAt:        now.Add(-time.Minute),
		},
	} {
		link, _, err := New(Dependencies{}).importedResourceRecords(input)
		if err != nil {
			t.Fatalf("build source link: %v", err)
		}
		if err := store.SaveImportSourceLink(ctx, link); err != nil {
			t.Fatalf("seed source link: %v", err)
		}
	}
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	application := New(Dependencies{
		Authorizer:    &fakeAuthorizer{},
		Tenants:       store,
		Inventories:   store,
		Audit:         store,
		ImportSources: source,
		ImportJobs:    store,
		ImportLinks:   store,
		IDs:           &fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
		Clock:         fakeClock{now: now},
	})

	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create import preview: %v", err)
	}
	if job.Counts.AssetsSkipped != 1 || job.Counts.AttachmentsSkipped != 1 || job.Counts.Warnings != 2 {
		t.Fatalf("expected preview duplicate skip counts, got %+v", job.Counts)
	}
	expected := map[string]importjob.Message{
		"duplicate-source-asset": {
			Code:       "duplicate-source-asset",
			Severity:   importjob.MessageSeverityWarning,
			Summary:    "Asset appears to have already been imported",
			Detail:     "source link already exists",
			SourceID:   "source:drill",
			SourceName: "Cordless drill",
		},
		"duplicate-source-attachment": {
			Code:       "duplicate-source-attachment",
			Severity:   importjob.MessageSeverityWarning,
			Summary:    "Attachment appears to have already been imported",
			Detail:     "source link already exists",
			SourceID:   "attachment:source:drill",
			SourceName: "drill.jpg",
		},
	}
	for _, message := range job.Preview.Messages {
		if want, ok := expected[message.Code]; ok {
			if message != want {
				t.Fatalf("unexpected duplicate message for %s: got %+v want %+v", message.Code, message, want)
			}
			delete(expected, message.Code)
		}
		encoded := message.Code + message.Summary + message.Detail + message.SourceID + message.SourceName
		if strings.Contains(encoded, "asset-existing") || strings.Contains(encoded, "attachment-existing") {
			t.Fatalf("duplicate message leaked persisted resource identifiers: %+v", message)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("expected asset and attachment duplicate messages, missing %+v from %+v", expected, job.Preview.Messages)
	}
}

func TestCreateImportJobPreviewReportsCSVAttachmentSourceLinkDuplicates(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	plan := importPlanForDurableJob("Homebox CSV", "source:drill")
	plan.Source = importplan.SourceSummary{Type: importplan.SourceLegacyHomeboxCSV, Name: "Homebox CSV", ImageImport: "unavailable"}
	plan.Fields = nil
	plan.Assets[0].CustomFields = map[string]any{}
	fingerprint, err := sourceFingerprint(plan)
	if err != nil {
		t.Fatalf("fingerprint csv plan: %v", err)
	}
	link, _, err := New(Dependencies{}).importedResourceRecords(importImportedResourceInput{
		TenantID:         tenant.ID("tenant-one"),
		InventoryID:      inventory.InventoryID("inventory-one"),
		JobID:            importjob.ID("previous-job"),
		SourceIdentity:   importSourceIdentity{sourceType: importplan.SourceLegacyHomeboxCSV, sourceInstanceKey: fingerprint},
		SourceEntityType: ports.ImportSourceEntityAttachment,
		SourceEntityID:   "attachment:source:drill",
		ResourceType:     ports.ImportResourceAttachment,
		ResourceID:       "attachment-existing",
		ResourceOwnerID:  "asset-existing",
		CreatedAt:        now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatalf("build csv source link: %v", err)
	}
	if err := store.SaveImportSourceLink(ctx, link); err != nil {
		t.Fatalf("seed csv source link: %v", err)
	}
	application := New(Dependencies{
		Authorizer:    &fakeAuthorizer{},
		Tenants:       store,
		Inventories:   store,
		Audit:         store,
		ImportSources: &fakeImportSourceReader{plan: plan},
		ImportJobs:    store,
		ImportLinks:   store,
		IDs:           &fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
		Clock:         fakeClock{now: now},
	})

	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomeboxCSV),
			FileName:   "homebox.csv",
		},
	})
	if err != nil {
		t.Fatalf("create csv import preview: %v", err)
	}
	if job.Source.Fingerprint != fingerprint || job.Counts.AttachmentsSkipped != 1 {
		t.Fatalf("expected csv duplicate attachment skip with fingerprint identity, got source=%+v counts=%+v", job.Source, job.Counts)
	}
	if len(job.Preview.Messages) != 1 || job.Preview.Messages[0].Code != "duplicate-source-attachment" {
		t.Fatalf("expected csv attachment duplicate message, got %+v", job.Preview.Messages)
	}
}

func TestCreateImportJobPreviewFailsWhenSourceLinkLookupFails(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	failure := errors.New("link store unavailable")
	application := New(Dependencies{
		Authorizer:    &fakeAuthorizer{},
		Tenants:       store,
		Inventories:   store,
		Audit:         store,
		ImportSources: &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")},
		ImportJobs:    store,
		ImportLinks:   failingImportLinkRepository{err: failure},
		IDs:           &fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
		Clock:         fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	_, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if !errors.Is(err, failure) {
		t.Fatalf("expected source-link lookup failure to abort preview, got %v", err)
	}
}

type failingImportLinkRepository struct {
	err error
}

func (f failingImportLinkRepository) ImportSourceLinkByKey(context.Context, ports.ImportSourceLinkKey) (ports.ImportSourceLink, bool, error) {
	return ports.ImportSourceLink{}, false, f.err
}

func (failingImportLinkRepository) SaveImportSourceLink(context.Context, ports.ImportSourceLink) error {
	return nil
}

func (failingImportLinkRepository) SaveImportJobResource(context.Context, ports.ImportJobResource) error {
	return nil
}

func (failingImportLinkRepository) ListImportJobResources(context.Context, tenant.ID, inventory.InventoryID, importjob.ID, ports.ImportJobResourcePageRequest) ([]ports.ImportJobResource, error) {
	return nil, nil
}

func (failingImportLinkRepository) ListAllImportJobResources(context.Context, tenant.ID, inventory.InventoryID, importjob.ID) ([]ports.ImportJobResource, error) {
	return nil, nil
}

func (failingImportLinkRepository) DeleteImportSourceLinksForJob(context.Context, tenant.ID, inventory.InventoryID, importjob.ID) (int, error) {
	return 0, nil
}

var _ ports.ImportLinkRepository = failingImportLinkRepository{}
