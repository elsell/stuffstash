package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (s Store) GetPrintJobContent(ctx context.Context, scope printing.Scope, id printing.JobID, now time.Time) ([]byte, error) {
	model, err := printJobByID(s.db.WithContext(ctx), scope, id)
	if err != nil {
		return nil, err
	}
	if !now.Before(model.ArtifactExpiresAt) || len(model.ArtifactContent) == 0 {
		return nil, ports.ErrPrintJobNotFound
	}
	return model.ArtifactContent, nil
}
