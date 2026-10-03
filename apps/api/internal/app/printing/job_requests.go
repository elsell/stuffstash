package printing

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *JobService) Reprint(ctx context.Context, scope LabelScope, predecessor p.JobID, key string, selection JobSelection) (p.Job, bool, error) {
	if err := s.access(ctx, scope, true); err != nil {
		return p.Job{}, false, err
	}
	existing, fingerprint, err := s.jobs.FindPrintJobRequest(ctx, jobScope(scope), string(scope.Principal.ID), key)
	if err == nil {
		expected := requestFingerprint(CreateJobInput{Scope: scope, AssetID: asset.ID(existing.AssetID), IdempotencyKey: key, Selection: selection, Predecessor: predecessor, Kind: existing.Kind})
		if existing.Predecessor != predecessor || fingerprint != expected {
			return p.Job{}, false, apperrors.ErrConflict
		}
		return existing, false, nil
	}
	if !errors.Is(err, ports.ErrPrintJobNotFound) {
		return p.Job{}, false, jobError(err)
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
