package printing

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

// prepareJob captures renderer output once, before any transactional writes.
func (s *JobService) prepareJob(ctx context.Context, scope LabelScope, printer printing.Printer, selection JobSelection, request printing.RenderRequest, assetID, labelID, key string) (printing.Job, []byte, error) {
	rendered, err := s.labels.deps.Renderer.Render(ctx, request)
	if err != nil || len(rendered.Content) == 0 || len(rendered.Content) > s.config.MaxArtifactBytes || s.config.ArtifactTTL <= 0 {
		return printing.Job{}, nil, apperrors.ErrInvalidInput
	}
	now := s.labels.deps.Clock.Now()
	id := printing.JobID(s.labels.deps.IDs.NewID())
	job := printing.Job{
		ID: id, Scope: jobScope(scope), PrinterID: printer.ID, Kind: printing.JobAssetLabel,
		AssetID: assetID, LabelReference: labelID, RequestedBy: string(scope.Principal.ID), IdempotencyKey: key,
		Media: printer.Media, MediaFingerprint: printer.MediaFingerprint, Template: request.Template, Content: request.Content,
		Copies: selection.Copies, Status: printing.JobQueued, Revision: 1, CreatedAt: now, UpdatedAt: now,
		Artifact: printing.Artifact{Key: string(id), SHA256: rendered.SHA256, ContentType: rendered.ContentType,
			ByteLength: int64(len(rendered.Content)), WidthPixels: rendered.WidthPixels, HeightPixels: rendered.HeightPixels, ExpiresAt: now.Add(s.config.ArtifactTTL)},
	}
	return job, rendered.Content, nil
}
