package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/homebox"
	"github.com/stuffstash/stuff-stash/internal/adapters/importworker"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/worklimit"
	"github.com/stuffstash/stuff-stash/internal/app"
	mediaapp "github.com/stuffstash/stuff-stash/internal/app/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func newSeededMediaTestApp(t *testing.T, state seededState, directUploads ports.DirectAttachmentUploader, imageProcessor ports.ImageProcessor) app.App {
	t.Helper()

	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	seedMemoryStore(t, context.Background(), store, authorizer, state)
	if fakeDirectUploads, ok := directUploads.(*httpFakeDirectAttachmentUploader); ok {
		fakeDirectUploads.blobs = store
	}

	var thumbnailReader ports.ThumbnailReader
	if batch, ok := imageProcessor.(ports.ImageBatchProcessor); ok {
		guard, err := memory.NewThumbnailPublicationGuard(store, ports.SystemClock{})
		if err != nil {
			t.Fatal(err)
		}
		processor, err := mediaapp.NewProcessor(store, store, batch, guard, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		admission, err := worklimit.New(1)
		if err != nil {
			t.Fatal(err)
		}
		thumbnailReader, err = mediaapp.NewReader(processor, admission, 250*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
	}
	application := app.New(app.Dependencies{
		Observer:                  &fakeObserver{},
		Auth:                      auth.NewLocalDevAuthenticator(),
		InvitationPublicBaseURL:   "https://stash.example.test/invitations/accept",
		Authorizer:                authorizer,
		Users:                     store,
		Tenants:                   store,
		TenantUnitOfWork:          store,
		Inventories:               store,
		InventoryUnitOfWork:       store,
		InventoryAccess:           store,
		InventoryAccessUnitOfWork: store,
		CustomAssetTypes:          store,
		CustomAssetTypeUnitOfWork: store,
		CustomFields:              store,
		CustomFieldUnitOfWork:     store,
		Assets:                    store,
		AssetTags:                 store,
		Checkouts:                 store,
		AssetUnitOfWork:           store,
		AssetTagUnitOfWork:        store,
		Undoables:                 store,
		Search:                    store,
		Attachments:               store,
		AttachmentUnitOfWork:      store,
		Blobs:                     store,
		DirectUploads:             directUploads,
		ImageProcessor:            imageProcessor,
		ThumbnailReader:           thumbnailReader,
		BlobDeletionOutbox:        store,
		Audit:                     store,
		Outbox:                    store,
		ProviderProfiles:          store,
		ProviderProfileUnitOfWork: store,
		VoiceProviderConfigs:      store,
		ProviderCredentialVault:   httpTestCredentialVault{repository: store, sealer: httpTestCredentialSealer{}},
		ProviderProfileTester:     httpTestProviderProfileTester{},
		RealtimeSessions:          store,
		ImportSources:             homebox.NewLegacyImporter(nil),
		ImportJobs:                store,
		ImportSourceVault:         newHTTPTestImportSourceVault(store),
		ImportLinks:               store,
		ImportAssetUnitOfWork:     store,
		IDs:                       &fakeIDGenerator{ids: state.ids},
	})
	return application.WithImportWorker(importworker.NewInProcess(application, nil))
}
