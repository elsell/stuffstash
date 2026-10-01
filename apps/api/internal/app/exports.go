package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
)

type ExportInventoryInput = dataportability.ExportInput
type ExportInventoryResult = dataportability.ExportResult

func (a App) ExportInventory(ctx context.Context, input ExportInventoryInput) (ExportInventoryResult, error) {
	return a.exportService.Export(ctx, input)
}
