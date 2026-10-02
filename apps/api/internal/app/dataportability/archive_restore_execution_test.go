package dataportability_test

import (
	"bytes"
	"context"
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
	for _, corrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "cross-instance restore", true: "changed preview rejected"}[corrupt], func(t *testing.T) {
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
			auth := memory.NewAuthorizer()
			principal := identity.Principal{ID: "new-owner"}
			must(auth.GrantTenantOwner(ctx, principal, "destination"))
			storage := blobstore.NewFileSystemStore(t.TempDir())
			deps := dataportability.ArchiveDependencies{Jobs: store, Commands: store, Authorizer: auth, Inventories: store, Tenants: store, IDs: &archiveTestIDs{}, Clock: archiveTestClock{}, Storage: storage, Scratch: blobstore.ScratchSpace{Directory: t.TempDir()}, MaxArchiveBytes: 1 << 20, Retention: time.Hour, CleanupTimeout: time.Second}
			service, err := dataportability.NewArchiveService(deps)
			must(err)
			publisher := gormstore.NewArchiveRestorePublisher(store, archiveTestClock{}, 100)
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
			if complete.State != archivejob.Ready {
				t.Fatal("restore not complete")
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
			stream, _, err := storage.OpenBlobStream(ctx, attachments[0].StorageKey)
			must(err)
			defer stream.Close()
			restored, err := io.ReadAll(stream)
			must(err)
			if !bytes.Equal(original, restored) {
				t.Fatal("cross-instance original changed")
			}
		})
	}
}
