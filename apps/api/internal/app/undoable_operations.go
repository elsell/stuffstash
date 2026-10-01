package app

import (
	"context"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
)

type ApplyUndoableOperationInput = assetapp.ApplyUndoableOperationInput

func (a App) UndoOperation(ctx context.Context, input ApplyUndoableOperationInput) (asset.Asset, error) {
	return a.assetService.UndoOperation(ctx, input)
}
func (a App) RedoOperation(ctx context.Context, input ApplyUndoableOperationInput) (asset.Asset, error) {
	return a.assetService.RedoOperation(ctx, input)
}
