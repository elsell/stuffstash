package printing

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func (s *JobService) Reprint(ctx context.Context, scope LabelScope, predecessor p.JobID, key string, selection JobSelection) (p.Job, bool, error) {
	if err := s.access(ctx, scope, true); err != nil {
		return p.Job{}, false, err
	}
	prior, err := s.jobs.GetPrintJob(ctx, jobScope(scope), predecessor)
	if err != nil {
		return p.Job{}, false, jobError(err)
	}
	if !prior.Terminal() || key == prior.IdempotencyKey {
		return p.Job{}, false, apperrors.ErrConflict
	}
	return s.Create(ctx, CreateJobInput{Scope: scope, AssetID: asset.ID(prior.AssetID), IdempotencyKey: key, Selection: selection, Predecessor: prior.ID, Kind: prior.Kind})
}

func (s *JobService) TestPrinter(ctx context.Context, scope LabelScope, key string, selection JobSelection) (p.Job, bool, error) {
	return s.Create(ctx, CreateJobInput{Scope: scope, IdempotencyKey: key, Selection: selection, Kind: p.JobPrinterTest})
}
