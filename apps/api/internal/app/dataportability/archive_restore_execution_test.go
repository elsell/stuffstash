package dataportability_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/inventoryarchive"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// A fresh database, authorization store, and blob root model another instance.
func verifyArchiveRestoreOnFreshInstance(t *testing.T, archive []byte, original []byte) {
	t.Helper()
	for _, scenario := range []string{"cross-instance restore", "changed preview rejected", "revoked finalization"} {
		t.Run(scenario, func(t *testing.T) {
			corrupt := scenario == "changed preview rejected"
			ctx := context.Background()
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			db, err := gormstore.OpenSQLite(filepath.Join(t.TempDir(), "destination.db"))
			must(err)
			must(gormstore.Migrate(ctx, db))
			store := gormstore.NewStore(db)
			must(store.SaveTenant(ctx, tenant.Tenant{ID: "destination", Name: "Other home"}))
			auth := &archiveFinalizationAuthorizer{Authorizer: memory.NewAuthorizer()}
			principal := identity.Principal{ID: "new-owner"}
			must(auth.GrantTenantOwner(ctx, principal, "destination"))
			storage := blobstore.NewFileSystemStore(t.TempDir())
			clock := &archiveAdvancingClock{}
			deps := dataportability.ArchiveDependencies{Jobs: store, Artifacts: store, Audit: store, Plans: inventoryarchive.PlanCodec{}, MaxMetadataBytes: 1 << 16, MaxRecords: 100, Commands: store, Authorizer: auth, Inventories: store, Tenants: store, IDs: &archiveTestIDs{}, Clock: clock, Storage: storage, Scratch: blobstore.ScratchSpace{Directory: t.TempDir()}, MaxArchiveBytes: 1 << 20, Retention: time.Hour, CleanupTimeout: time.Second}
			service, err := dataportability.NewArchiveService(deps)
			must(err)
			publisher := gormstore.NewArchiveRestorePublisher(store, clock, 100)
			worker, err := dataportability.NewArchiveWorker(dataportability.ArchiveWorkerDependencies{Service: service, Snapshots: store, Metadata: inventoryarchive.MetadataCodec{}, Packages: inventoryarchive.PackageCodec{}, Readers: inventoryarchive.PackageCodec{}, Plans: inventoryarchive.PlanCodec{}, Fields: store, Types: store, Publisher: publisher, Limits: ports.ArchivePackageLimits{CompressedBytes: 1 << 20, ExpandedBytes: 1 << 20, MetadataBytes: 1 << 16, EntryBytes: 1 << 20, Entries: 100}, MaxRecords: 100, LeaseDuration: time.Minute, HeartbeatInterval: time.Second})
			must(err)
			access := dataportability.ArchiveAccess{Principal: principal, TenantID: "destination"}
			job, err := service.UploadRestore(ctx, access, "restore", bytes.NewReader(archive))
			must(err)
			must(worker.RunJob(ctx, job))
			preview, err := service.Job(ctx, access, job.ID)
			must(err)
			if preview.State != archivejob.AwaitingApproval || preview.DestinationInventoryID == "inventory" {
				t.Fatal("missing independent preview")
			}
			if _, found, err := store.InventoryByID(ctx, "destination", inventory.InventoryID(preview.DestinationInventoryID)); err != nil || found {
				t.Fatal("inventory published before approval")
			}
			if corrupt {
				key, ok := media.NewStorageKey(preview.PlanArtifactID)
				if !ok {
					t.Fatal("key")
				}
				must(storage.PutBlobStream(ctx, ports.BlobStreamWrite{Key: key, ContentType: "application/json", SizeBytes: 2, MaxBytes: 1 << 16, Content: bytes.NewReader([]byte("{}"))}))
			}
			approved, err := service.Approve(ctx, access, job.ID, "Restored camera")
			must(err)
			err = worker.RunJob(ctx, approved)
			if corrupt {
				if err == nil {
					t.Fatal("changed plan accepted")
				}
				if _, found, err := store.InventoryByID(ctx, "destination", inventory.InventoryID(preview.DestinationInventoryID)); err != nil || found {
					t.Fatal("changed plan published")
				}
				return
			}
			must(err)
			complete, err := service.Job(ctx, access, job.ID)
			must(err)
			if complete.Phase != archivejob.Finalization || complete.State != archivejob.Queued {
				t.Fatal("restore reported complete before grant")
			}
			must(worker.RunJob(ctx, complete))
			pending, err := service.Job(ctx, access, job.ID)
			must(err)
			if pending.State != archivejob.Queued || pending.OwnerGrantEventID != complete.OwnerGrantEventID {
				t.Fatal("pending grant lost publication")
			}
			events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "grant-worker", 10, clock.Now(), clock.Now().Add(time.Minute))
			must(err)
			for _, event := range events {
				must(auth.GrantInventoryOwner(ctx, principal, event.TenantID, event.InventoryID))
				must(store.MarkAuthorizationOutboxEventProcessed(ctx, event.ID, event.ClaimID))
			}
			eligible, err := store.ListRunnableArchiveJobs(ctx, clock.Now(), 100)
			must(err)
			if len(eligible) != 0 {
				t.Fatal("pending finalization remains immediately eligible")
			}
			clock.offset.Add(int64(time.Second))
			auth.deny = scenario == "revoked finalization"
			err = worker.RunJob(ctx, pending)
			if auth.deny {
				if !errors.Is(err, ports.ErrForbidden) {
					t.Fatal("revoked restore finalized", err)
				}
			} else {
				must(err)
			}
			complete, err = service.Job(ctx, access, job.ID)
			must(err)
			expected := archivejob.Ready
			if auth.deny {
				expected = archivejob.Failed
			}
			if auth.deny && complete.Failure != archivejob.FailurePermission {
				t.Fatal("permission denial misclassified", complete.Failure)
			}
			if complete.State != expected {
				t.Fatal("incorrect finalization state", complete.State)
			}

			destination := inventory.InventoryID(complete.DestinationInventoryID)
			inv, found, err := store.InventoryByID(ctx, "destination", destination)
			must(err)
			if !found || inv.Name != "Restored camera" {
				t.Fatal("approved name missing")
			}
			assets, err := store.ListAssetsByInventory(ctx, "destination", destination, ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll, Sort: ports.AssetListSortIDAsc})
			must(err)
			if len(assets) != 1 || assets[0].ID == asset.ID("asset") || assets[0].Title != "Camera" {
				t.Fatal("asset not remapped")
			}
			attachments, err := store.ListAttachmentsByAsset(ctx, "destination", destination, assets[0].ID, ports.AttachmentListPageRequest{Limit: 10, IncludeArchived: true})
			must(err)
			if len(attachments) != 1 || attachments[0].ID == "photo" {
				t.Fatal("attachment not remapped")
			}
			now := (archiveTestClock{}).Now()
			thumbnails, err := store.ClaimThumbnailJobs(ctx, "restored-thumbnails", 10, now, now.Add(time.Minute))
			must(err)
			if len(thumbnails) != 1 || thumbnails[0].Job.AttachmentID != attachments[0].ID {
				t.Fatal("restored image missing thumbnail work")
			}

			stream, _, err := storage.OpenBlobStream(ctx, attachments[0].StorageKey)
			must(err)
			defer stream.Close()
			restored, err := io.ReadAll(stream)
			must(err)
			if !bytes.Equal(original, restored) {
				t.Fatal("cross-instance original changed")
			}
			stream.Close()
			clock.offset.Store(int64(2 * time.Hour))
			must(service.CleanupArchives(ctx, 100))
			expired, err := service.Job(ctx, access, job.ID)
			must(err)
			if expired.State != archivejob.Expired {
				t.Fatal("retention did not expire job")
			}
			deletions, err := store.ClaimPendingBlobDeletionEvents(ctx, "archive-cleanup", 100, clock.Now(), clock.Now().Add(time.Minute))
			must(err)
			if len(deletions) != 2 {
				t.Fatalf("expected source and plan cleanup, got %d", len(deletions))
			}
			for _, event := range deletions {
				if event.StorageKey == attachments[0].StorageKey {
					t.Fatal("cleanup deletes restored original")
				}
				must(storage.DeleteBlob(ctx, event.StorageKey))
				must(store.MarkBlobDeletionEventProcessed(ctx, event.ID, event.ClaimID))
			}
			kept, _, err := storage.OpenBlobStream(ctx, attachments[0].StorageKey)
			must(err)
			kept.Close()

		})
	}
}

type archiveFinalizationAuthorizer struct {
	ports.Authorizer
	deny bool
}

func (a *archiveFinalizationAuthorizer) CheckInventory(ctx context.Context, p identity.Principal, permission ports.InventoryPermission, id inventory.InventoryID) error {
	if a.deny {
		return ports.ErrForbidden
	}
	return a.Authorizer.CheckInventory(ctx, p, permission, id)
}
