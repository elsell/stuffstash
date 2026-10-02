package dataportability_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type archiveTestClock struct{}

func (archiveTestClock) Now() time.Time { return time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC) }

type archiveTestIDs struct{ n int }

func (g *archiveTestIDs) NewID() string { g.n++; return "archive-" + strconv.Itoa(g.n) }
func TestArchiveJobsRequireOwnerAndCurrentScopePermission(t *testing.T) {
	ctx := context.Background()
	db, err := gormstore.OpenSQLite(filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err = gormstore.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	store := gormstore.NewStore(db)
	if err = store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: "Home"}); err != nil {
		t.Fatal(err)
	}
	if err = store.SaveInventory(ctx, inventory.Inventory{ID: "inventory", TenantID: "home", Name: "Inventory", LifecycleState: inventory.LifecycleStateActive}); err != nil {
		t.Fatal(err)
	}
	authorizer := memory.NewAuthorizer()
	owner := identity.Principal{ID: "owner"}
	other := identity.Principal{ID: "other"}
	if err = authorizer.GrantTenantOwner(ctx, owner, "home"); err != nil {
		t.Fatal(err)
	}
	if err = authorizer.GrantInventoryOwner(ctx, owner, "home", "inventory"); err != nil {
		t.Fatal(err)
	}
	if err = authorizer.GrantInventoryViewer(ctx, other, "home", "inventory"); err != nil {
		t.Fatal(err)
	}
	commands := &commitAcknowledgmentFailure{ArchiveJobCommands: store}
	blobRoot := t.TempDir()
	service, err := dataportability.NewArchiveService(dataportability.ArchiveDependencies{Jobs: store, Artifacts: store, Commands: commands, Authorizer: authorizer, Inventories: store, Tenants: store, IDs: &archiveTestIDs{}, Clock: archiveTestClock{}, Storage: blobstore.NewFileSystemStore(blobRoot), Scratch: blobstore.ScratchSpace{Directory: t.TempDir()}, MaxArchiveBytes: 1024, Retention: time.Hour, CleanupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	access := dataportability.ArchiveAccess{Principal: owner, TenantID: "home", InventoryID: "inventory"}
	job, err := service.CreateExport(ctx, access, "request", true, true)
	if err != nil {
		t.Fatal(err)
	}
	again, err := service.CreateExport(ctx, access, "request", true, true)
	if err != nil || again.ID != job.ID {
		t.Fatalf("idempotency: %v", err)
	}
	wrong := access
	wrong.Principal = other
	if _, err = service.Job(ctx, wrong, job.ID); err == nil {
		t.Fatal("other inventory viewer accessed private archive")
	}
	listed, err := service.List(ctx, wrong, "", 10)
	if err != nil || len(listed) != 0 {
		t.Fatal("private job leaked through list")
	}
	if _, err = service.Cancel(ctx, wrong, job.ID); err == nil {
		t.Fatal("other principal cancelled archive")
	}
	cross := access
	cross.TenantID = "elsewhere"
	if _, err = service.Job(ctx, cross, job.ID); err == nil {
		t.Fatal("cross-tenant job access")
	}
	restoreAccess := dataportability.ArchiveAccess{Principal: other, TenantID: "home"}
	if _, err = service.UploadRestore(ctx, restoreAccess, "restore", bytes.NewReader([]byte("archive"))); err == nil {
		t.Fatal("viewer started restore")
	}
	restoreAccess.Principal = owner
	restore, err := service.UploadRestore(ctx, restoreAccess, "restore", bytes.NewReader([]byte("archive")))
	if err != nil {
		t.Fatal(err)
	}
	same, err := service.UploadRestore(ctx, restoreAccess, "restore", bytes.NewReader([]byte("archive")))
	if err != nil || same.ID != restore.ID {
		t.Fatalf("same upload retry: %v", err)
	}
	if _, err = service.UploadRestore(ctx, restoreAccess, "restore", bytes.NewReader([]byte("different"))); err == nil {
		t.Fatal("different source replaced upload")
	}
	if _, err = service.UploadRestore(ctx, restoreAccess, "oversized", bytes.NewReader(make([]byte, 1025))); err == nil {
		t.Fatal("upload limit ignored")
	}
	files := 0
	if err = filepath.WalkDir(blobRoot, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files++
		}
		return nil
	}); err != nil || files != 2 {
		t.Fatalf("expected source plus retained uncertain upload: %d %v", files, err)
	}
	commands.fail = true
	if _, err = service.UploadRestore(ctx, restoreAccess, "lost-ack", bytes.NewReader([]byte("durable source"))); err == nil {
		t.Fatal("expected lost commit acknowledgment")
	}
	persisted, found, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home", PrincipalID: "owner"}, commands.committed.ID)
	if err != nil || !found {
		t.Fatalf("committed job missing: %v", err)
	}
	storageKey, ok := media.NewStorageKey(persisted.SourceArtifactID)
	if !ok {
		t.Fatal("invalid artifact key")
	}
	stream, _, err := blobstore.NewFileSystemStore(blobRoot).OpenBlobStream(ctx, storageKey)
	if err != nil {
		t.Fatalf("committed job lost source: %v", err)
	}
	data, err := io.ReadAll(stream)
	stream.Close()
	if err != nil || string(data) != "durable source" {
		t.Fatalf("source changed: %q %v", data, err)
	}
	commands.fail = false
	recovered, err := service.UploadRestore(ctx, restoreAccess, "lost-ack", bytes.NewReader([]byte("durable source")))
	if err != nil || recovered.ID != persisted.ID {
		t.Fatalf("retry lost job: %v", err)
	}
	viewerJob, err := service.CreateExport(ctx, wrong, "viewer-request", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = authorizer.RevokeInventoryViewer(ctx, other, "home", "inventory"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Job(ctx, wrong, viewerJob.ID); err == nil {
		t.Fatal("revoked permission retained job access")
	}
	if _, err = service.Approve(ctx, restoreAccess, restore.ID, "New inventory"); err == nil {
		t.Fatal("unvalidated upload approved")
	}
	if _, err = service.Cancel(ctx, access, job.ID); err != nil {
		t.Fatal(err)
	}
}

// Models a committed transaction whose acknowledgment is lost in transport.
type commitAcknowledgmentFailure struct {
	ports.ArchiveJobCommands
	fail      bool
	committed archivejob.Record
}

func (f *commitAcknowledgmentFailure) CreateArchiveJobAudited(ctx context.Context, job archivejob.Record, record audit.Record) (archivejob.Record, error) {
	result, err := f.ArchiveJobCommands.CreateArchiveJobAudited(ctx, job, record)
	if err == nil && f.fail {
		f.committed = result
		return archivejob.Record{}, errors.New("commit acknowledgment lost")
	}
	return result, err
}
