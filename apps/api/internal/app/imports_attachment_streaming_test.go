package app

import (
	"context"
	"errors"
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
