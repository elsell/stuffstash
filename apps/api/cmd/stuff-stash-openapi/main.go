package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/homebox"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	store := memory.NewStore()
	application := app.New(app.Dependencies{
		Observer:                  observability.NewFanOut(),
		Auth:                      auth.NewLocalDevAuthenticator(),
		Authorizer:                memory.NewAuthorizer(),
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
		Checkouts:                 store,
		AssetUnitOfWork:           store,
		Undoables:                 store,
		Search:                    store,
		Attachments:               store,
		AttachmentUnitOfWork:      store,
		Blobs:                     store,
		BlobDeletionOutbox:        store,
		Audit:                     store,
		Outbox:                    store,
		ImportSources:             homebox.NewLegacyImporter(nil),
	})
	server := httpserver.NewServer(":0", application)
	request := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request.WithContext(context.Background()))
	if response.Code != http.StatusOK {
		return fmt.Errorf("generate openapi: expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if _, err := os.Stdout.Write(response.Body.Bytes()); err != nil {
		return fmt.Errorf("write openapi: %w", err)
	}
	return nil
}
