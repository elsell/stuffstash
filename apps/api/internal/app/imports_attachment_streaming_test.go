package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteImportJobReadsAttachmentBytesOneAtATimeThroughSourcePort(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Fields = nil
	plan.Assets[0].CustomFields = map[string]any{}
	plan.Attachments[0].FileName = "drill.png"
	plan.Attachments[0].ContentType = "image/png"
	plan.Attachments[0].Content = pngAttachmentBytes()
	plan.Attachments[0].Content = nil
	plan.Attachments[0].SizeBytes = 0
	plan.Attachments = append(plan.Attachments, importplan.Attachment{
		SourceID:      "attachment:source:drill:second",
		AssetSourceID: "source:drill",
		FileName:      "drill-side.png",
		ContentType:   "image/png",
	})
	source := &fakeImportSourceReader{plan: plan}
	attachmentSource := &recordingImportAttachmentSource{
		contentBySourceID: map[string]ports.ImportAttachmentContent{
			"attachment:source:drill": {
				FileName:    "drill.png",
				ContentType: "image/png",
				Content:     pngAttachmentBytes(),
			},
			"attachment:source:drill:second": {
				FileName:    "drill-side.png",
				ContentType: "image/png",
				Content:     pngAttachmentBytes(),
			},
		},
	}
	application := New(Dependencies{
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   store,
		Inventories:               store,
		CustomAssetTypes:          store,
		CustomAssetTypeUnitOfWork: store,
		CustomFields:              store,
		CustomFieldUnitOfWork:     store,
		Assets:                    store,
		AssetUnitOfWork:           store,
		Undoables:                 store,
		Audit:                     store,
		AttachmentUnitOfWork:      store,
		Attachments:               store,
		Blobs:                     store,
		BlobDeletionOutbox:        store,
		ImportSources:             source,
		ImportAttachmentSources:   attachmentSource,
		ImportJobs:                store,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:                   store,
		ImportAssetUnitOfWork:         store,
		ImportAttachmentUnitOfWork:    store,
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset",
			"attachment-one", "audit-attachment-one", "attachment-two", "audit-attachment-two", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AttachmentsCreated != 2 {
		t.Fatalf("expected two streamed attachments, got %+v", result)
	}
	if got := attachmentSource.readSourceIDs; len(got) != 2 || got[0] != "attachment:source:drill" || got[1] != "attachment:source:drill:second" {
		t.Fatalf("attachment reads = %#v", got)
	}
	if attachmentSource.openCount != 1 {
		t.Fatalf("attachment sessions opened = %d", attachmentSource.openCount)
	}
}

func TestImportAttachmentReadFailureMessagePreservesSafeCause(t *testing.T) {
	attachment := importplan.Attachment{SourceID: "attachment-one", FileName: "photo.jpg"}
	tests := map[string]struct {
		err         error
		wantCode    string
		wantSummary string
	}{
		"download": {
			err:         ports.NewImportAttachmentReadError(ports.ImportAttachmentDownloadFailed, errors.New("provider detail")),
			wantCode:    "attachment-unavailable",
			wantSummary: "Attachment could not be downloaded",
		},
		"oversized": {
			err:         ports.NewImportAttachmentReadError(ports.ImportAttachmentTooLarge, errors.New("provider detail")),
			wantCode:    "attachment-too-large",
			wantSummary: "Attachment is too large",
		},
		"unsupported": {
			err:         ports.NewImportAttachmentReadError(ports.ImportAttachmentUnsupportedType, errors.New("provider detail")),
			wantCode:    "attachment-unsupported-type",
			wantSummary: "Attachment type is not supported",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			message := importAttachmentReadFailureMessage(test.err, attachment)
			if message.Code != test.wantCode || message.Summary != test.wantSummary {
				t.Fatalf("message = %+v", message)
			}
			if message.SourceID != attachment.SourceID || message.SourceName != attachment.FileName {
				t.Fatalf("message lost safe source identity: %+v", message)
			}
		})
	}
}

func TestImportAttachmentSessionFailureMessageRetainsSafeActionableDetail(t *testing.T) {
	tests := map[string]struct {
		err        error
		wantDetail string
	}{
		"provider status": {
			err:        ports.NewImportSourceUserError("Homebox returned 429 Too Many Requests"),
			wantDetail: "Homebox returned 429 Too Many Requests",
		},
		"timeout": {
			err:        context.DeadlineExceeded,
			wantDetail: "The source did not respond before image import timed out",
		},
		"internal detail is redacted": {
			err:        errors.New("dial tcp 192.168.2.96:443: connection refused with token secret"),
			wantDetail: "Stuff Stash could not establish a source session",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			message := importAttachmentSessionFailureMessage(test.err)
			if message.Code != "attachment-session-unavailable" || message.Severity != importplan.SeverityError {
				t.Fatalf("message classification = %+v", message)
			}
			if !strings.Contains(message.Detail, test.wantDetail) || !strings.Contains(message.Detail, "Already imported records were kept") {
				t.Fatalf("message detail = %q", message.Detail)
			}
			if strings.Contains(message.Detail, "192.168.2.96") || strings.Contains(message.Detail, "token secret") {
				t.Fatalf("message leaked internal detail: %q", message.Detail)
			}
		})
	}
}

type failingOpenImportAttachmentSource struct {
	err error
}

func (f failingOpenImportAttachmentSource) OpenImportAttachmentSession(context.Context, ports.ImportSourceRequest) (ports.ImportAttachmentSession, error) {
	return nil, f.err
}

func TestExecuteImportJobPersistsAttachmentSessionDiagnosticBeforeCredentialCleanup(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Fields = nil
	plan.Assets[0].CustomFields = map[string]any{}
	source := &fakeImportSourceReader{plan: plan}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
	application := New(Dependencies{
		Authorizer: &fakeAuthorizer{}, Tenants: store, Inventories: store,
		CustomAssetTypes: store, CustomAssetTypeUnitOfWork: store,
		CustomFields: store, CustomFieldUnitOfWork: store,
		Assets: store, AssetUnitOfWork: store, Undoables: store, Audit: store,
		AttachmentUnitOfWork: store, Attachments: store, Blobs: store, BlobDeletionOutbox: store,
		ImportSources: source,
		ImportAttachmentSources: failingOpenImportAttachmentSource{
			err: ports.NewImportSourceUserError("Homebox returned 429 Too Many Requests"),
		},
		ImportJobs: store, ImportSourceVault: vault, ImportLinks: store,
		ImportAssetUnitOfWork: store, ImportAttachmentUnitOfWork: store,
		ImportWorker: &fakeImportWorker{}, BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "audit-failed", "audit-credential-cleaned",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusFailed || result.Counts.AssetsCreated != 1 || result.Counts.AttachmentsCreated != 0 {
		t.Fatalf("failed import counts = %+v", result)
	}
	if result.Progress.Phase != importjob.PhaseTerminal || result.Progress.Done != 0 || result.Progress.Total != len(plan.Attachments) {
		t.Fatalf("terminal attachment progress = %+v", result.Progress)
	}
	if len(result.Messages) != 1 || result.Messages[0].Code != "attachment-session-unavailable" || !strings.Contains(result.Messages[0].Detail, "429 Too Many Requests") {
		t.Fatalf("durable diagnostics = %+v", result.Messages)
	}
	if _, retained := vault.requests[result.ID]; retained {
		t.Fatal("terminal import credentials were not cleaned after diagnostic persistence")
	}
}

func TestExecuteImportJobPersistsAttachmentStorageDiagnostic(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Fields = nil
	plan.Assets[0].CustomFields = map[string]any{}
	plan.Attachments[0].FileName = "drill.png"
	plan.Attachments[0].ContentType = "image/png"
	plan.Attachments[0].Content = pngAttachmentBytes()
	source := &fakeImportSourceReader{plan: plan}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
	application := New(Dependencies{
		Authorizer: &fakeAuthorizer{}, Tenants: store, Inventories: store,
		CustomAssetTypes: store, CustomAssetTypeUnitOfWork: store,
		CustomFields: store, CustomFieldUnitOfWork: store,
		Assets: store, AssetUnitOfWork: store, Undoables: store, Audit: store,
		AttachmentUnitOfWork: store, Attachments: store, Blobs: failingBlobStorage{}, BlobDeletionOutbox: store,
		ImportSources: source, ImportAttachmentSources: source,
		ImportJobs: store, ImportSourceVault: vault, ImportLinks: store,
		ImportAssetUnitOfWork: store, ImportAttachmentUnitOfWork: store,
		ImportWorker: &fakeImportWorker{}, BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "attachment-one", "audit-attachment", "audit-failed", "audit-credential-cleaned",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusFailed || result.Progress.Done != 0 || result.Progress.Total != len(plan.Attachments) {
		t.Fatalf("failed storage import = %+v", result)
	}
	if len(result.Messages) != 1 || result.Messages[0].Code != "attachment-storage-unavailable" || result.Messages[0].SourceName != "drill.png" {
		t.Fatalf("durable storage diagnostic = %+v", result.Messages)
	}
	if strings.Contains(result.Messages[0].Detail, "storage unavailable") {
		t.Fatalf("storage diagnostic leaked adapter error: %+v", result.Messages)
	}
}
