package dataportability_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/inventoryarchive"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestArchiveWorkerExportsOriginalAndFencesPublication(t *testing.T) {
	for _, scenario := range []string{"ready", "permission revoked", "cancelled", "lease expired", "reclaimed"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			db, err := gormstore.OpenSQLite(filepath.Join(t.TempDir(), "worker.db"))
			if err != nil {
				t.Fatal(err)
			}
			if err = gormstore.Migrate(ctx, db); err != nil {
				t.Fatal(err)
			}
			store := gormstore.NewStore(db)
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			must(store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: "Home"}))
			must(store.SaveInventory(ctx, inventory.Inventory{ID: "inventory", TenantID: "home", Name: "Inventory", LifecycleState: inventory.LifecycleStateActive}))
			clock := &archiveAdvancingClock{}
			now := clock.Now()
			entry := asset.Asset{ID: "asset", TenantID: "home", InventoryID: "inventory", Title: "Camera", Kind: asset.KindItem, LifecycleState: asset.LifecycleStateActive, CreatedAt: now, UpdatedAt: now}
			record, ok := audit.NewRecord("created", "home", "inventory", "viewer", audit.ActionAssetCreated, audit.SourceAPI, audit.TargetAsset, "asset", now, "", nil)
			if !ok {
				t.Fatal("audit")
			}
			must(store.CreateAsset(ctx, entry, record, nil))
			content := []byte("original image bytes")
			hash := fmt.Sprintf("%x", sha256.Sum256(content))
			attachment := media.Attachment{ID: "photo", TenantID: "home", InventoryID: "inventory", AssetID: "asset", StorageKey: "home/inventory/asset/photo", FileName: "photo.jpg", ContentType: media.ContentTypeJPEG, SizeBytes: int64(len(content)), SHA256: media.SHA256(hash), CreatedAt: now, LifecycleState: media.LifecycleStateActive}
			record.ID = "attached"
			record.Action = audit.ActionAttachmentCreated
			record.TargetType = audit.TargetAttachment
			record.TargetID = "photo"
			thumbnail, thumbnailErr := media.PlanThumbnailJob(attachment)
			must(thumbnailErr)
			must(store.SaveAttachment(ctx, attachment, record, thumbnail))
			storage := &archiveCompletionStorage{StreamingBlobStorage: blobstore.NewFileSystemStore(t.TempDir())}
			must(storage.PutBlobStream(ctx, ports.BlobStreamWrite{Key: attachment.StorageKey, ContentType: "image/jpeg", SizeBytes: int64(len(content)), MaxBytes: 1024, Content: bytes.NewReader(content)}))
			auth := memory.NewAuthorizer()
			principal := identity.Principal{ID: "viewer"}
			must(auth.GrantInventoryViewer(ctx, principal, "home", "inventory"))
			deps := dataportability.ArchiveDependencies{Jobs: store, Artifacts: store, Audit: store, Plans: inventoryarchive.PlanCodec{}, MaxMetadataBytes: 1 << 16, MaxRecords: 100, Commands: store, Authorizer: auth, Inventories: store, Tenants: store, IDs: &archiveTestIDs{}, Clock: clock, Storage: storage, Scratch: blobstore.ScratchSpace{Directory: t.TempDir()}, MaxArchiveBytes: 1 << 20, Retention: time.Hour, CleanupTimeout: time.Second}
			service, err := dataportability.NewArchiveService(deps)
			must(err)
			access := dataportability.ArchiveAccess{Principal: principal, TenantID: "home", InventoryID: "inventory"}
			job, err := service.CreateExport(ctx, access, "export", true, false)
			must(err)
			storage.after = func() error {
				switch scenario {
				case "permission revoked":
					return auth.RevokeInventoryViewer(ctx, principal, "home", "inventory")
				case "lease expired", "reclaimed":
					clock.offset.Store(int64(2 * time.Minute))
					if scenario == "reclaimed" {
						prior, found, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home", SourceInventoryID: "inventory"}, job.ID)
						if err != nil {
							return err
						}
						if !found {
							return fmt.Errorf("missing job")
						}
						next, err := prior.Claim("replacement", clock.Now(), clock.Now().Add(time.Minute))
						if err != nil {
							return err
						}
						changed, err := store.UpdateArchiveJob(ctx, next, prior.Revision)
						if err != nil {
							return err
						}
						if !changed {
							return fmt.Errorf("claim lost")
						}
					}
					return nil
				case "cancelled":
					_, err := service.Cancel(ctx, access, job.ID)
					return err
				}
				return nil
			}
			worker, err := dataportability.NewArchiveWorker(dataportability.ArchiveWorkerDependencies{Service: service, Snapshots: store, Metadata: inventoryarchive.MetadataCodec{}, Packages: inventoryarchive.PackageCodec{}, Readers: inventoryarchive.PackageCodec{}, Plans: inventoryarchive.PlanCodec{}, Fields: store, Types: store, Publisher: gormstore.NewArchiveRestorePublisher(store, clock, 100), Limits: ports.ArchivePackageLimits{CompressedBytes: 1 << 20, ExpandedBytes: 1 << 20, MetadataBytes: 1 << 16, EntryBytes: 1 << 20, Entries: 100}, MaxRecords: 100, LeaseDuration: time.Minute, HeartbeatInterval: time.Second})
			must(err)
			err = worker.RunJob(ctx, job)
			current, found, readErr := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home", SourceInventoryID: "inventory"}, job.ID)
			must(readErr)
			if !found {
				t.Fatal("lost job")
			}
			history, historyErr := store.ListInventoryAuditRecords(ctx, "home", "inventory", ports.AuditRecordPageRequest{Limit: 100})
			must(historyErr)
			hasFinalAudit := false
			for _, event := range history {
				if event.TargetID == job.ID && event.Action == audit.ActionArchiveJobUpdated {
					hasFinalAudit = true
				}
			}
			if (scenario == "ready" || scenario == "permission revoked") && !hasFinalAudit {
				t.Fatal("missing final job audit")
			}
			if scenario != "ready" {
				if err == nil {
					t.Fatal("lost authority reported success")
				}
				if scenario == "permission revoked" && (current.State != archivejob.Failed || current.Failure != archivejob.FailurePermission) {
					t.Fatal("permission failure not recorded")
				}
				if scenario == "cancelled" && current.State != archivejob.Cancelled {
					t.Fatal("cancellation overwritten")
				}
				if scenario == "reclaimed" && current.LeaseToken != "replacement" {
					t.Fatal("replacement lease overwritten")
				}
				if current.State == archivejob.Ready || current.ResultArtifactID != "" {
					t.Fatal("stale work published")
				}
				return
			}
			must(err)
			if current.State != archivejob.Ready {
				t.Fatalf("state %s", current.State)
			}
			key, ok := media.NewStorageKey(current.ResultArtifactID)
			if !ok {
				t.Fatal("key")
			}
			stream, size, err := storage.OpenBlobStream(ctx, key)
			must(err)
			defer stream.Close()
			archive, err := inventoryarchive.Read(ctx, stream, size, inventoryarchive.Limits{CompressedBytes: 1 << 20, ExpandedBytes: 1 << 20, MetadataBytes: 1 << 16, EntryBytes: 1 << 20, Entries: 100})
			must(err)
			metadata, err := (inventoryarchive.MetadataCodec{}).DecodeMetadata(ctx, archive.InventoryJSON, 1<<16)
			must(err)
			if len(metadata.Assets) != 1 || metadata.Assets[0].Title != "Camera" || len(archive.Media) != 1 || archive.Selection.OtherFiles {
				t.Fatal("incomplete archive")
			}
			original, err := archive.Open(hash)
			must(err)
			defer original.Close()
			restored, err := io.ReadAll(original)
			must(err)
			if !bytes.Equal(restored, content) {
				t.Fatal("original lost")
			}
			packageBytes := make([]byte, size)
			_, readErr = stream.ReadAt(packageBytes, 0)
			must(readErr)
			verifyArchiveRestoreOnFreshInstance(t, packageBytes, content)
		})
	}
}

type archiveCompletionStorage struct {
	ports.StreamingBlobStorage
	after func() error
}

func (s *archiveCompletionStorage) PutBlobStream(ctx context.Context, w ports.BlobStreamWrite) error {
	if err := s.StreamingBlobStorage.PutBlobStream(ctx, w); err != nil {
		return err
	}
	if s.after != nil && strings.HasPrefix(w.Key.String(), "archives/") {
		return s.after()
	}
	return nil
}

// Atomic time control lets storage completion cross a lease boundary safely.
type archiveAdvancingClock struct{ offset atomic.Int64 }

func (c *archiveAdvancingClock) Now() time.Time {
	return (archiveTestClock{}).Now().Add(time.Duration(c.offset.Load()))
}
