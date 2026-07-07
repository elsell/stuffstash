package gormstore

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestImportJobRepositoryPersistsScopedJobHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	job := gormImportJob("job-one", time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	got, found, err := store.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID)
	if err != nil {
		t.Fatalf("read import job: %v", err)
	}
	if !found {
		t.Fatalf("expected import job to be found")
	}
	if got.Source.Fingerprint != job.Source.Fingerprint || got.Progress.Phase != importjob.PhaseReady || len(got.Messages) != 1 {
		t.Fatalf("unexpected persisted job: %+v", got)
	}
	if got.Source.AllowPrivateNetwork || got.Source.AllowInsecureTLS {
		t.Fatalf("expected CSV import job source options to default false, got %+v", got.Source)
	}
	if _, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-other"), inventory.InventoryID(job.InventoryID.String()), job.ID); err != nil || found {
		t.Fatalf("expected wrong tenant read to miss, found=%t err=%v", found, err)
	}

	jobs, err := store.ListImportJobs(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), nilPage())
	if err != nil {
		t.Fatalf("list import jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("expected job in history, got %+v", jobs)
	}

	liveJob := gormImportJob("job-live", job.CreatedAt.Add(time.Minute))
	liveJob.Source.Type = importjob.SourceTypeLegacyHomebox
	liveJob.Source.Name = "Homebox"
	liveJob.Source.BaseURL = "https://homebox.example.test"
	liveJob.Source.ImageImport = "enabled"
	liveJob.Source.AllowPrivateNetwork = true
	liveJob.Source.AllowInsecureTLS = true
	if err := store.SaveImportJob(ctx, liveJob); err != nil {
		t.Fatalf("save live import job: %v", err)
	}
	liveGot, found, err := store.ImportJobByID(ctx, tenant.ID(liveJob.TenantID.String()), inventory.InventoryID(liveJob.InventoryID.String()), liveJob.ID)
	if err != nil || !found {
		t.Fatalf("read live import job found=%t err=%v", found, err)
	}
	if !liveGot.Source.AllowPrivateNetwork || !liveGot.Source.AllowInsecureTLS {
		t.Fatalf("expected safe live source options to round trip, got %+v", liveGot.Source)
	}
}

func TestImportJobRepositoryUpdatesAndHidesRemovedHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	job := gormImportJob("job-one", time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}

	job.Status = importjob.StatusCancelRequested
	job.CancellationMode = importjob.CancellationModeDiscardPartial
	job.CancellationRequestID = "cancel-request-one"
	job.Progress.Message = "Cancellation requested"
	job.UpdatedAt = job.UpdatedAt.Add(time.Minute)
	if err := store.UpdateImportJob(ctx, job); err != nil {
		t.Fatalf("update import job: %v", err)
	}
	updated, found, err := store.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID)
	if err != nil {
		t.Fatalf("read updated import job: %v", err)
	}
	if !found ||
		updated.CancellationMode != importjob.CancellationModeDiscardPartial ||
		updated.CancellationRequestID != "cancel-request-one" ||
		updated.Progress.Message != "Cancellation requested" {
		t.Fatalf("unexpected updated job found=%t job=%+v", found, updated)
	}

	job.Status = importjob.StatusCancelledDiscarded
	removedAt := job.UpdatedAt.Add(time.Minute)
	removed, err := store.MarkImportJobHistoryRemoved(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, removedAt, job.UpdatedAt)
	if err != nil {
		t.Fatalf("remove import job from history: %v", err)
	}
	if !removed {
		t.Fatalf("expected remove import job from history to update")
	}
	if _, found, err := store.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID); err != nil || found {
		t.Fatalf("expected removed job detail to miss, found=%t err=%v", found, err)
	}
	jobs, err := store.ListImportJobs(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), nilPage())
	if err != nil {
		t.Fatalf("list after remove: %v", err)
	}
	if len(jobs) != 0 {
		t.Fatalf("expected removed job hidden from history, got %+v", jobs)
	}
	stale := job
	stale.Status = importjob.StatusSucceeded
	stale.UpdatedAt = removedAt.Add(time.Minute)
	if err := store.UpdateImportJob(ctx, stale); err == nil {
		t.Fatalf("expected stale generic update of removed job to fail")
	}
	if _, found, err := store.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID); err != nil || found {
		t.Fatalf("expected stale update not to resurrect removed job, found=%t err=%v", found, err)
	}
	recoverable, err := store.ListImportJobsByStatus(ctx, ports.ImportJobStatusPageRequest{Status: importjob.StatusCancelledDiscarded, Limit: 10})
	if err != nil {
		t.Fatalf("list by status after remove: %v", err)
	}
	if len(recoverable) != 0 {
		t.Fatalf("expected removed job hidden from status recovery list, got %+v", recoverable)
	}
}

func TestImportJobRepositoryConditionallyTransitionsAndClaimsJobs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := gormImportJob("job-one", now)
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}

	running := job
	running.Status = importjob.StatusRunning
	running.UpdatedAt = now.Add(time.Minute)
	running.Progress.UpdatedAt = running.UpdatedAt
	started, err := store.UpdateImportJobIfStatus(ctx, running, importjob.StatusPreviewed)
	if err != nil {
		t.Fatalf("conditional start: %v", err)
	}
	if !started {
		t.Fatalf("expected first conditional start to update")
	}
	staleStarted, err := store.UpdateImportJobIfStatus(ctx, running, importjob.StatusPreviewed)
	if err != nil {
		t.Fatalf("stale conditional start: %v", err)
	}
	if staleStarted {
		t.Fatalf("expected stale conditional start to be rejected")
	}

	claimed := running
	claimed.UpdatedAt = running.UpdatedAt.Add(time.Minute)
	claimed.Progress.UpdatedAt = claimed.UpdatedAt
	ok, err := store.ClaimImportJob(ctx, claimed, running.UpdatedAt)
	if err != nil {
		t.Fatalf("claim import job: %v", err)
	}
	if !ok {
		t.Fatalf("expected first claim to update")
	}
	staleClaim, err := store.ClaimImportJob(ctx, claimed, running.UpdatedAt)
	if err != nil {
		t.Fatalf("stale claim import job: %v", err)
	}
	if staleClaim {
		t.Fatalf("expected stale claim to be rejected")
	}
}

func TestImportJobRepositoryConditionallyUpdatesProgress(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := gormImportJob("job-one", now)
	job.Status = importjob.StatusRunning
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}

	first := importjob.Progress{Phase: importjob.PhaseAssets, Done: 1, Total: 2, Message: "Creating assets", UpdatedAt: now.Add(time.Minute)}
	updated, err := store.UpdateImportJobProgress(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, first, job.UpdatedAt)
	if err != nil {
		t.Fatalf("progress update: %v", err)
	}
	if !updated {
		t.Fatalf("expected progress update to apply")
	}
	stale := importjob.Progress{Phase: importjob.PhaseAssets, Done: 2, Total: 2, Message: "Creating assets", UpdatedAt: now.Add(2 * time.Minute)}
	staleUpdated, err := store.UpdateImportJobProgress(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, stale, job.UpdatedAt)
	if err != nil {
		t.Fatalf("stale progress update: %v", err)
	}
	if staleUpdated {
		t.Fatalf("expected stale progress update to be rejected")
	}
	got, found, err := store.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID)
	if err != nil || !found {
		t.Fatalf("read import job found=%t err=%v", found, err)
	}
	if got.Progress.Done != 1 || got.Progress.Total != 2 || got.UpdatedAt != first.UpdatedAt {
		t.Fatalf("unexpected persisted progress after stale update: %+v", got)
	}
	if len(got.ProgressHistory) != 2 || got.ProgressHistory[0].Phase != importjob.PhaseReady || got.ProgressHistory[1].Phase != importjob.PhaseAssets {
		t.Fatalf("expected progress phase history to persist ready and assets, got %+v", got.ProgressHistory)
	}

	nonAdvancing := importjob.Progress{Phase: importjob.PhaseAssets, Done: 2, Total: 2, Message: "Creating assets", UpdatedAt: first.UpdatedAt.Add(time.Nanosecond)}
	if _, err := store.UpdateImportJobProgress(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, nonAdvancing, first.UpdatedAt); !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected non-advancing microsecond progress timestamp to be rejected, got %v", err)
	}

	second := importjob.Progress{Phase: importjob.PhaseAssets, Done: 2, Total: 2, Message: "Creating assets", UpdatedAt: now.Add(3 * time.Minute)}
	updated, err = store.UpdateImportJobProgress(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, second, first.UpdatedAt)
	if err != nil {
		t.Fatalf("second progress update: %v", err)
	}
	if !updated {
		t.Fatalf("expected second progress update to apply")
	}
	got, found, err = store.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID)
	if err != nil || !found {
		t.Fatalf("read import job after second progress found=%t err=%v", found, err)
	}
	if got.Progress.Done != 2 {
		t.Fatalf("expected current progress to advance, got %+v", got.Progress)
	}
	if len(got.ProgressHistory) != 2 {
		t.Fatalf("expected same-phase progress to keep bounded phase history, got %+v", got.ProgressHistory)
	}
	if got.ProgressHistory[1].Done != 2 || got.ProgressHistory[1].Total != 2 || !got.ProgressHistory[1].UpdatedAt.Equal(second.UpdatedAt) {
		t.Fatalf("expected bounded phase history to refresh latest safe progress counts, got %+v", got.ProgressHistory[1])
	}
}

func TestImportLinkRepositoryEnforcesSourceUniqueness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	job := gormImportJob("job-one", time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
	job.Source.Type = importjob.SourceTypeLegacyHomebox
	job.Source.BaseURL = "https://homebox.example.test"
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}

	link := ports.ImportSourceLink{
		Key: ports.ImportSourceLinkKey{
			TenantID:          tenant.ID(job.TenantID.String()),
			InventoryID:       inventory.InventoryID(job.InventoryID.String()),
			SourceType:        importplan.SourceLegacyHomebox,
			SourceInstanceKey: "https://homebox.example.test",
			SourceEntityType:  ports.ImportSourceEntityAsset,
			SourceEntityID:    "source:drill",
		},
		ResourceType: ports.ImportResourceAsset,
		ResourceID:   "asset-one",
		JobID:        job.ID,
		CreatedAt:    job.CreatedAt,
	}
	if err := store.SaveImportSourceLink(ctx, link); err != nil {
		t.Fatalf("save source link: %v", err)
	}
	if err := store.SaveImportSourceLink(ctx, link); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate source link conflict, got %v", err)
	}
	got, found, err := store.ImportSourceLinkByKey(ctx, link.Key)
	if err != nil {
		t.Fatalf("read source link: %v", err)
	}
	if !found || got.ResourceID != "asset-one" {
		t.Fatalf("unexpected source link found=%t link=%+v", found, got)
	}
}

func TestCreateImportedAttachmentRollsBackOnSourceLinkConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	item := assetItem("asset-one", "tenant-home", "inventory-home", asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("save asset: %v", err)
	}
	job := gormImportJob("job-one", now)
	job.Source.Type = importjob.SourceTypeLegacyHomebox
	job.Source.BaseURL = "https://homebox.example.test"
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	link := ports.ImportSourceLink{
		Key: ports.ImportSourceLinkKey{
			TenantID:          tenant.ID(job.TenantID.String()),
			InventoryID:       inventory.InventoryID(job.InventoryID.String()),
			SourceType:        importplan.SourceLegacyHomebox,
			SourceInstanceKey: "https://homebox.example.test",
			SourceEntityType:  ports.ImportSourceEntityAttachment,
			SourceEntityID:    "attachment:source:drill",
		},
		ResourceType: ports.ImportResourceAttachment,
		ResourceID:   "attachment-one",
		JobID:        job.ID,
		CreatedAt:    now,
	}
	if err := store.SaveImportSourceLink(ctx, link); err != nil {
		t.Fatalf("save existing source link: %v", err)
	}
	attachment := gormImportedAttachment(t, now)
	resource := ports.ImportJobResource{
		TenantID:          tenant.ID(job.TenantID.String()),
		InventoryID:       inventory.InventoryID(job.InventoryID.String()),
		JobID:             job.ID,
		ResourceType:      ports.ImportResourceAttachment,
		ResourceID:        attachment.ID.String(),
		ResourceOwnerID:   item.ID.String(),
		SourceType:        link.Key.SourceType,
		SourceInstanceKey: link.Key.SourceInstanceKey,
		SourceEntityType:  link.Key.SourceEntityType,
		SourceEntityID:    link.Key.SourceEntityID,
		CreatedAt:         now,
	}
	err := store.CreateImportedAttachment(ctx, attachment, auditRecord(t, "audit-attachment", tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), audit.ActionAttachmentCreated), link, resource)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected source link conflict, got %v", err)
	}
	if _, found, err := store.AttachmentByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), item.ID, attachment.ID); err != nil || found {
		t.Fatalf("expected attachment rollback, found=%t err=%v", found, err)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), ports.AuditRecordPageRequest{Limit: 20})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	for _, record := range records {
		if record.ID == audit.ID("audit-attachment") {
			t.Fatalf("expected attachment audit record rollback, got %+v", record)
		}
	}
	resources, err := store.ListImportJobResources(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, ports.ImportJobResourcePageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list import job resources: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("expected import job resource rollback, got %+v", resources)
	}
}

func TestImportJobResourceListIsBoundedAndOrdered(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := gormImportJob("job-one", now)
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	for index := 0; index < 60; index++ {
		record := ports.ImportJobResource{
			TenantID:          tenant.ID(job.TenantID.String()),
			InventoryID:       inventory.InventoryID(job.InventoryID.String()),
			JobID:             job.ID,
			ResourceType:      ports.ImportResourceAsset,
			ResourceID:        fmt.Sprintf("asset-%02d", index),
			SourceType:        importplan.SourceLegacyHomeboxCSV,
			SourceInstanceKey: "sha256:test",
			SourceEntityType:  ports.ImportSourceEntityAsset,
			SourceEntityID:    fmt.Sprintf("source-%02d", index),
			CreatedAt:         now.Add(time.Duration(59-index) * time.Second),
		}
		if err := store.SaveImportJobResource(ctx, record); err != nil {
			t.Fatalf("save import job resource %d: %v", index, err)
		}
	}

	resources, err := store.ListImportJobResources(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID, ports.ImportJobResourcePageRequest{Limit: 12})
	if err != nil {
		t.Fatalf("list import job resources: %v", err)
	}
	if len(resources) != 12 {
		t.Fatalf("expected bounded import job resources, got %d", len(resources))
	}
	if resources[0].ResourceID != "asset-59" || resources[11].ResourceID != "asset-48" {
		t.Fatalf("expected resources ordered at repository boundary, got first=%s last=%s", resources[0].ResourceID, resources[11].ResourceID)
	}
}

func TestImportJobSourceVacuumDeletesTerminalJobSourcesBeforeExpiry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := gormImportJob("job-one", now)
	job.Status = importjob.StatusSucceeded
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	if err := store.ReplaceImportJobSource(ctx, ports.ImportJobSourceRecord{
		Scope: ports.ImportJobSourceScope{
			TenantID:    tenant.ID(job.TenantID.String()),
			InventoryID: inventory.InventoryID(job.InventoryID.String()),
			JobID:       job.ID,
		},
		Sealed: ports.SealedImportJobSource{
			KeyID:      "key-one",
			Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
			Nonce:      []byte("123456789012"),
			Ciphertext: []byte("sealed-source"),
		},
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("save import source: %v", err)
	}
	deleted, err := store.DeleteVacuumableImportJobSources(ctx, []importjob.Status{importjob.StatusSucceeded}, now)
	if err != nil {
		t.Fatalf("vacuum sources: %v", err)
	}
	if len(deleted) != 1 || deleted[0].JobID != job.ID {
		t.Fatalf("expected one deleted source scope, got %+v", deleted)
	}
	if _, found, err := store.ImportJobSource(ctx, ports.ImportJobSourceScope{TenantID: tenant.ID(job.TenantID.String()), InventoryID: inventory.InventoryID(job.InventoryID.String()), JobID: job.ID}); err != nil || found {
		t.Fatalf("expected source removed, found=%t err=%v", found, err)
	}
}

func gormImportedAttachment(t *testing.T, now time.Time) media.Attachment {
	t.Helper()
	attachment, ok := media.NewAttachment(
		media.ID("attachment-one"),
		media.TenantID("tenant-home"),
		media.InventoryID("inventory-home"),
		media.AssetID("asset-one"),
		media.StorageKey("tenant-home/inventory-home/asset-one/attachment-one"),
		media.FileName("drill.png"),
		media.ContentTypePNG,
		64,
		media.SHA256("0000000000000000000000000000000000000000000000000000000000000000"),
		now,
	)
	if !ok {
		t.Fatalf("expected valid imported attachment")
	}
	return attachment
}

func gormImportJob(id string, now time.Time) importjob.Record {
	return importjob.Record{
		ID:          importjob.ID(id),
		TenantID:    importjob.TenantID("tenant-home"),
		InventoryID: importjob.InventoryID("inventory-home"),
		ActorID:     importjob.PrincipalID("owner"),
		Status:      importjob.StatusPreviewed,
		Source: importjob.SourceRef{
			Type:        importjob.SourceTypeLegacyHomeboxCSV,
			Name:        "Homebox CSV",
			Version:     "0.24.0",
			ImageImport: "unavailable",
			Fingerprint: "sha256:test",
		},
		Counts: importjob.Counts{Fields: 1, Assets: 1, Warnings: 1},
		Messages: []importjob.Message{{
			Code:     "csv-images-unavailable",
			Severity: importjob.MessageSeverityWarning,
			Summary:  "Images are unavailable for CSV imports",
		}},
		Progress: importjob.Progress{
			Phase:     importjob.PhaseReady,
			Done:      2,
			Total:     2,
			Message:   "Preview ready",
			UpdatedAt: now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func nilPage() ports.ImportJobPageRequest {
	return ports.ImportJobPageRequest{}
}
