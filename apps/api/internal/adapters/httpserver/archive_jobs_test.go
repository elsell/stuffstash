package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/inventoryarchive"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestArchiveJobHTTPBoundary(t *testing.T) {
	runArchiveJobHTTPBoundary(t, nil, false)
}
func runArchiveJobHTTPBoundary(t *testing.T, coverage *executedScenarioCoverage, adversarial bool) {
	ctx := context.Background()
	clock := &archiveBoundaryClock{now: (ports.SystemClock{}).Now()}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	db, err := gormstore.OpenSQLite(filepath.Join(t.TempDir(), "archive.db"))
	must(err)
	must(gormstore.Migrate(ctx, db))
	store := gormstore.NewStore(db)
	must(store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: "Home"}))
	must(store.SaveTenant(ctx, tenant.Tenant{ID: "elsewhere", Name: "Elsewhere"}))
	must(store.SaveInventory(ctx, inventory.Inventory{ID: "inventory", TenantID: "home", Name: "Inventory", LifecycleState: inventory.LifecycleStateActive}))
	authorizer := memory.NewAuthorizer()
	owner := identity.Principal{ID: "owner"}
	viewer := identity.Principal{ID: "viewer"}
	must(authorizer.GrantTenantOwner(ctx, owner, "home"))
	must(authorizer.GrantInventoryOwner(ctx, owner, "home", "inventory"))
	must(authorizer.GrantInventoryViewer(ctx, viewer, "home", "inventory"))
	readAudit := &archiveReadAuditFailure{AuditRepository: store}
	service, err := dataportability.NewArchiveService(dataportability.ArchiveDependencies{Jobs: store, Artifacts: store, Commands: store, Audit: readAudit, Plans: inventoryarchive.PlanCodec{}, MaxMetadataBytes: 1 << 16, MaxRecords: 100, Authorizer: authorizer, Inventories: store, Tenants: store, IDs: idgen.NewULIDGenerator(), Clock: clock, Storage: blobstore.NewFileSystemStore(t.TempDir()), Scratch: blobstore.ScratchSpace{Directory: t.TempDir()}, MaxArchiveBytes: 1 << 20, Retention: time.Hour, CleanupTimeout: time.Second})
	must(err)
	server := NewServerWithOptions(":0", newTestAppWithAuthorizer(nil, authorizer), Options{Archives: &service, MaxJSONBodyBytes: 512})
	if coverage != nil {
		next := server.Handler
		server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := httptest.NewRecorder()
			next.ServeHTTP(response, r)
			if r.Pattern != "" && ((adversarial && response.Code >= 400) || (!adversarial && response.Code < 400)) {
				coverage.operation[r.Pattern] = struct{}{}
			}
			for name, values := range response.Header() {
				w.Header()[name] = values
			}
			w.WriteHeader(response.Code)
			_, _ = w.Write(response.Body.Bytes())
		})
	}
	path := "/tenants/home/archive-jobs"
	body := map[string]any{"inventoryId": "inventory", "photos": true, "otherFiles": true}
	headers := map[string]string{"Idempotency-Key": "request"}
	for _, token := range []string{"", "Bearer invalid", "Bearer dev:stranger"} {
		r := performRequestWithHeaders(server, http.MethodPost, path, token, headers, body)
		if r.Code != 401 && r.Code != 403 {
			t.Fatalf("unauthorized create %d %s", r.Code, r.Body.String())
		}
	}
	created := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:owner", headers, body)
	if created.Code != 201 {
		t.Fatalf("create %d %s", created.Code, created.Body.String())
	}
	var job struct {
		Data struct {
			ID    string `json:"id"`
			State string `json:"state"`
		}
	}
	must(json.Unmarshal(created.Body.Bytes(), &job))
	if job.Data.ID == "" || job.Data.State != "queued" {
		t.Fatal("missing job")
	}
	item := path + "/" + job.Data.ID
	all := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
	if all.Code != 200 || !strings.Contains(all.Body.String(), job.Data.ID) {
		t.Fatalf("household listing missing export: %d %s", all.Code, all.Body.String())
	}
	for _, test := range []struct {
		path, token string
		code        int
	}{{item + "?inventoryId=inventory", "Bearer dev:viewer", 404}, {item, "Bearer dev:owner", 200}, {"/tenants/elsewhere/archive-jobs/" + job.Data.ID + "?inventoryId=inventory", "Bearer dev:owner", 404}, {item + "?inventoryId=inventory", "", 401}} {
		r := performRequest(server, http.MethodGet, test.path, test.token, nil)
		if r.Code != test.code {
			t.Fatalf("private job %d want%d: %s", r.Code, test.code, r.Body.String())
		}
	}
	for _, attempt := range []struct{ method, path string }{{http.MethodGet, path + "?inventoryId=inventory"}, {http.MethodDelete, item + "?inventoryId=inventory"}, {http.MethodGet, item + "/preview?inventoryId=inventory"}} {
		denied := performRequest(server, attempt.method, attempt.path, "", nil)
		if denied.Code != 401 {
			t.Fatalf("unauthenticated archive operation: %d", denied.Code)
		}
	}
	listed := performRequest(server, http.MethodGet, path+"?inventoryId=inventory", "Bearer dev:viewer", nil)
	if listed.Code != 200 || strings.Contains(listed.Body.String(), job.Data.ID) {
		t.Fatal("private listing leaked")
	}
	for _, suffix := range []string{"/retry", "/approve"} {
		r := performRequest(server, http.MethodPost, item+suffix+"?inventoryId=inventory", "Bearer dev:viewer", map[string]any{"name": "Stolen"})
		if r.Code != 404 {
			t.Fatalf("other principal action: %s", r.Body.String())
		}
	}
	r := performRequest(server, http.MethodPost, item+"/approve?inventoryId=inventory", "Bearer dev:owner", map[string]any{"name": "Invalid"})
	if r.Code != 409 {
		t.Fatal("unvalidated export approved")
	}
	r = performRequest(server, http.MethodGet, item+"?inventoryId=inventory", "Bearer dev:owner", nil)
	if r.Code != 200 {
		t.Fatal(r.Body.String())
	}
	for _, private := range []string{"artifact", "leaseToken", "principalId", "sourceSHA256"} {
		if strings.Contains(r.Body.String(), private) {
			t.Fatal("private state exposed")
		}
	}
	// A real ZIP travels through the streaming boundary, independently of the
	// much smaller JSON request limit. Validation and publication run the real worker.
	downloadPath := item + "/content?inventoryId=inventory"
	pending := performRequest(server, http.MethodGet, downloadPath, "Bearer dev:owner", nil)
	if pending.Code != 409 {
		t.Fatalf("pending download: %d %s", pending.Code, pending.Body.String())
	}
	worker, err := dataportability.NewArchiveWorker(dataportability.ArchiveWorkerDependencies{Service: service, Snapshots: store, Metadata: inventoryarchive.MetadataCodec{}, Packages: inventoryarchive.PackageCodec{}, Readers: inventoryarchive.PackageCodec{}, Plans: inventoryarchive.PlanCodec{}, Fields: store, Types: store, Publisher: gormstore.NewArchiveRestorePublisher(store, clock, 100), Limits: ports.ArchivePackageLimits{CompressedBytes: 1 << 20, ExpandedBytes: 1 << 20, MetadataBytes: 1 << 16, EntryBytes: 1 << 20, Entries: 100}, MaxRecords: 100, LeaseDuration: time.Minute, HeartbeatInterval: time.Second})
	must(err)
	exportJob, found, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home", SourceInventoryID: "inventory"}, job.Data.ID)
	must(err)
	if !found {
		t.Fatal("missing export")
	}
	must(worker.RunJob(ctx, exportJob))
	download := performRequest(server, http.MethodGet, item+"/content", "Bearer dev:owner", nil)
	if download.Code != 200 || download.Header().Get("Content-Type") != "application/zip" || download.Body.Len() <= 512 {
		t.Fatalf("download: %d %s", download.Code, download.Body.String())
	}
	archive := append([]byte(nil), download.Body.Bytes()...)
	for _, token := range []string{"", "Bearer invalid", "Bearer dev:viewer"} {
		denied := performRequest(server, http.MethodGet, downloadPath, token, nil)
		if denied.Code != 401 && denied.Code != 404 {
			t.Fatalf("download leaked: %d", denied.Code)
		}
	}
	upload := func(token, key, contentType string, data []byte) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/tenants/home/archive-restores", bytes.NewReader(data))
		request.Header.Set("Authorization", token)
		request.Header.Set("Idempotency-Key", key)
		request.Header.Set("Content-Type", contentType)
		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		return response
	}
	for _, token := range []string{"", "Bearer invalid", "Bearer dev:viewer"} {
		denied := upload(token, "upload", "application/zip", archive)
		if denied.Code != 401 && denied.Code != 403 {
			t.Fatalf("upload role boundary: %d %s", denied.Code, denied.Body.String())
		}
	}
	for _, test := range []struct {
		key, contentType string
		data             []byte
		code             int
	}{{"", "application/zip", archive, 400}, {"wrong-type", "text/plain", archive, 415}, {"too-large", "application/zip", make([]byte, (1<<20)+1), 413}} {
		response := upload("Bearer dev:owner", test.key, test.contentType, test.data)
		if response.Code != test.code {
			t.Fatalf("upload validation: %d want %d %s", response.Code, test.code, response.Body.String())
		}
	}
	restored := upload("Bearer dev:owner", "upload", "application/zip", archive)
	if restored.Code != 201 {
		t.Fatalf("upload: %d %s", restored.Code, restored.Body.String())
	}
	var restore struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	must(json.Unmarshal(restored.Body.Bytes(), &restore))
	restorePath := path + "/" + restore.Data.ID
	premature := performRequest(server, http.MethodPost, restorePath+"/approve", "Bearer dev:owner", map[string]any{"name": "Restored"})
	if premature.Code != 409 {
		t.Fatal("unvalidated restore approved")
	}
	restoreJob, found, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
	must(err)
	if !found {
		t.Fatal("missing restore")
	}
	must(worker.RunJob(ctx, restoreJob))
	preview := performRequest(server, http.MethodGet, restorePath+"/preview", "Bearer dev:owner", nil)
	if preview.Code != 200 || !strings.Contains(preview.Body.String(), "Inventory") {
		t.Fatalf("preview: %d %s", preview.Code, preview.Body.String())
	}
	duplicate := upload("Bearer dev:owner", "upload", "application/zip", archive)
	if duplicate.Code != 201 || !strings.Contains(duplicate.Body.String(), restore.Data.ID) {
		t.Fatal("upload retry created another job")
	}
	approved := performRequest(server, http.MethodPost, restorePath+"/approve", "Bearer dev:owner", map[string]any{"name": "Restored"})
	if approved.Code != 200 {
		t.Fatalf("approve: %d %s", approved.Code, approved.Body.String())
	}
	restoreJob, _, err = store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
	must(err)
	must(worker.RunJob(ctx, restoreJob))
	publishedStatus := performRequest(server, http.MethodGet, restorePath, "Bearer dev:owner", nil)
	if publishedStatus.Code != 200 || !strings.Contains(publishedStatus.Body.String(), "finalization") || strings.Contains(publishedStatus.Body.String(), `"state":"ready"`) {
		t.Fatalf("published restore exposed ready: %s", publishedStatus.Body.String())
	}
	cancelPublished := performRequest(server, http.MethodDelete, restorePath, "Bearer dev:owner", nil)
	if cancelPublished.Code != http.StatusConflict {
		t.Fatal("cancelled published restore")
	}
	restoreJob, _, err = store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
	must(err)
	must(worker.RunJob(ctx, restoreJob)) // delayed outbox must not complete or republish
	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "restore-grants", 100, clock.Now(), clock.Now().Add(time.Minute))
	must(err)
	for _, event := range events {
		if event.ID != restoreJob.OwnerGrantEventID {
			continue
		}
		must(authorizer.GrantInventoryOwner(ctx, owner, event.TenantID, event.InventoryID))
		must(store.MarkAuthorizationOutboxEventProcessed(ctx, event.ID, event.ClaimID))
	}
	restoreJob, _, err = store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
	must(err)
	clock.now = clock.now.Add(time.Second)
	must(worker.RunJob(ctx, restoreJob))
	completed := performRequest(server, http.MethodGet, restorePath, "Bearer dev:owner", nil)
	if completed.Code != 200 || !strings.Contains(completed.Body.String(), "ready") {
		t.Fatalf("restore status: %s", completed.Body.String())
	}
	// Revoking access also revokes an already completed archive download.
	viewerExport, err := service.CreateExport(ctx, dataportability.ArchiveAccess{Principal: viewer, TenantID: "home", InventoryID: "inventory"}, "viewer-export", false, false)
	must(err)
	must(worker.RunJob(ctx, viewerExport))
	must(authorizer.RevokeInventoryViewer(ctx, viewer, "home", "inventory"))
	denied := performRequest(server, http.MethodGet, path+"/"+viewerExport.ID+"/content?inventoryId=inventory", "Bearer dev:viewer", nil)
	if denied.Code != 403 {
		t.Fatalf("revoked download: %d", denied.Code)
	}
	queued := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:owner", map[string]string{"Idempotency-Key": "cancel-me"}, body)
	must(json.Unmarshal(queued.Body.Bytes(), &job))
	newest := performRequest(server, http.MethodGet, path+"?inventoryId=inventory&limit=1", "Bearer dev:owner", nil)
	var page struct {
		Data []struct {
			ID string `json:"id"`
		}
		Meta struct {
			Pagination struct {
				NextCursor string `json:"nextCursor"`
			}
		}
	}
	must(json.Unmarshal(newest.Body.Bytes(), &page))
	if len(page.Data) != 1 || page.Data[0].ID != job.Data.ID {
		t.Fatal("archive jobs not newest first")
	}
	older := performRequest(server, http.MethodGet, path+"?inventoryId=inventory&limit=1&after="+page.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	must(json.Unmarshal(older.Body.Bytes(), &page))
	if len(page.Data) != 1 || page.Data[0].ID == job.Data.ID {
		t.Fatal("archive cursor repeated latest job")
	}
	item = path + "/" + job.Data.ID
	r = performRequest(server, http.MethodDelete, item+"?inventoryId=inventory", "Bearer dev:owner", nil)
	if r.Code != 200 || !strings.Contains(r.Body.String(), "cancelled") {
		t.Fatal("cancel failed")
	}
	// User-controlled ZIP paths never become filesystem destinations or an
	// approvable restore, even for a principal allowed to upload archives.
	for index, name := range []string{"../escape", "/absolute", `..\escape`, `C:\escape`} {
		t.Run("malicious ZIP "+name, func(t *testing.T) {
			payload := appendArchivePath(t, archive, name)
			response := upload("Bearer dev:owner", fmt.Sprintf("path-attack-%d", index), "application/zip", payload)
			if response.Code != http.StatusCreated {
				t.Fatal(response.Body.String())
			}
			must(json.Unmarshal(response.Body.Bytes(), &restore))
			candidate, found, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
			must(err)
			if !found {
				t.Fatal("missing quarantined job")
			}
			if err = worker.RunJob(ctx, candidate); err == nil {
				t.Fatal("malicious archive validated")
			}
			status := performRequest(server, http.MethodGet, path+"/"+restore.Data.ID, "Bearer dev:owner", nil)
			if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), `"state":"failed"`) {
				t.Fatalf("unsafe status: %s", status.Body.String())
			}
			approval := performRequest(server, http.MethodPost, path+"/"+restore.Data.ID+"/approve", "Bearer dev:owner", map[string]any{"name": "Unsafe"})
			if approval.Code != http.StatusConflict {
				t.Fatalf("unsafe approval: %d", approval.Code)
			}
			candidate, _, err = store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
			must(err)
			if candidate.DestinationInventoryID != "" {
				t.Fatal("unsafe destination reserved")
			}
		})
	}
	// A corrupt archive reaches a durable validation failure and can be retried.
	invalid := upload("Bearer dev:owner", "invalid-archive", "application/zip", []byte("not a ZIP"))
	if invalid.Code != 201 {
		t.Fatal(invalid.Body.String())
	}
	must(json.Unmarshal(invalid.Body.Bytes(), &restore))
	failedJob, _, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "home"}, restore.Data.ID)
	must(err)
	if err = worker.RunJob(ctx, failedJob); err == nil {
		t.Fatal("invalid archive accepted")
	}
	retried := performRequest(server, http.MethodPost, path+"/"+restore.Data.ID+"/retry", "Bearer dev:owner", nil)
	if retried.Code != 200 || !strings.Contains(retried.Body.String(), "queued") {
		t.Fatalf("retry: %d %s", retried.Code, retried.Body.String())
	}
	// Household discovery remains creator-private and rechecks revoked source access.
	must(store.SaveInventory(ctx, inventory.Inventory{ID: "other", TenantID: "home", Name: "Other", LifecycleState: inventory.LifecycleStateActive}))
	must(authorizer.GrantInventoryViewer(ctx, viewer, "home", "other"))
	must(authorizer.GrantInventoryViewer(ctx, viewer, "home", "inventory"))
	otherExport := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:viewer", map[string]string{"Idempotency-Key": "other-export"}, map[string]any{"inventoryId": "other", "photos": false, "otherFiles": false})
	if otherExport.Code != 201 {
		t.Fatal(otherExport.Body.String())
	}
	var otherJob struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	must(json.Unmarshal(otherExport.Body.Bytes(), &otherJob))
	ownExport := performRequestWithHeaders(server, http.MethodPost, path, "Bearer dev:viewer", map[string]string{"Idempotency-Key": "viewer-discovery-export"}, body)
	if ownExport.Code != 201 {
		t.Fatal(ownExport.Body.String())
	}
	var viewerJob struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	must(json.Unmarshal(ownExport.Body.Bytes(), &viewerJob))
	visible := performRequest(server, http.MethodGet, path, "Bearer dev:viewer", nil)
	if visible.Code != 200 || !strings.Contains(visible.Body.String(), viewerJob.Data.ID) || strings.Contains(visible.Body.String(), job.Data.ID) {
		t.Fatalf("creator isolation: %s", visible.Body.String())
	}
	must(authorizer.RevokeInventoryViewer(ctx, viewer, "home", "inventory"))
	hidden := performRequest(server, http.MethodGet, path+"?limit=1", "Bearer dev:viewer", nil)
	if hidden.Code != 200 || strings.Contains(hidden.Body.String(), viewerJob.Data.ID) || !strings.Contains(hidden.Body.String(), otherJob.Data.ID) {
		t.Fatalf("revoked export leaked: %s", hidden.Body.String())
	}
	deniedJob := performRequest(server, http.MethodGet, path+"/"+viewerJob.Data.ID, "Bearer dev:viewer", nil)
	if deniedJob.Code != 404 {
		t.Fatalf("revoked job: %d", deniedJob.Code)
	}
	readAudit.fail = true
	for _, path := range []string{downloadPath, item + "?inventoryId=inventory", restorePath + "/preview", "/tenants/home/archive-jobs?inventoryId=inventory"} {
		response := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
		if response.Code != 500 || response.Header().Get("Content-Type") == "application/zip" {
			t.Fatalf("audit failure disclosed archive data: %d", response.Code)
		}
	}
	readAudit.fail = false
	history, err := store.ListInventoryAuditRecords(ctx, "home", "inventory", ports.AuditRecordPageRequest{Limit: 100})
	must(err)
	foundReadAudit := false
	for _, record := range history {
		if record.Action.String() == "archive_job.viewed" && record.Metadata["operation"] == "download" {
			foundReadAudit = true
		}
	}
	if !foundReadAudit {
		t.Fatal("read audit missing")
	}
}

// A controlled repository failure verifies the public read boundary fails closed.
type archiveReadAuditFailure struct {
	ports.AuditRepository
	fail bool
}

func (a *archiveReadAuditFailure) SaveAuditRecord(ctx context.Context, record audit.Record) error {
	if a.fail {
		return errors.New("audit repository unavailable")
	}
	return a.AuditRepository.SaveAuditRecord(ctx, record)
}

type archiveBoundaryClock struct{ now time.Time }

func (c *archiveBoundaryClock) Now() time.Time { return c.now }
