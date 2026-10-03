package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) WithAssetPrintUnitOfWork(uow ports.AssetPrintUnitOfWork) App {
	a.assetPrintUnitOfWork = uow
	return a
}
func (a App) CreateAssetAndPrint(ctx context.Context, input CreateAssetInput, selection printing.JobSelection, key string) (ports.AssetPrintResult, error) {
	if a.printJobs == nil {
		return ports.AssetPrintResult{}, apperrors.ErrInvalidInput
	}
	return a.printJobs.CreateAssetAndPrint(ctx, a.assetService, a.assetPrintUnitOfWork, input, selection, key)
}
