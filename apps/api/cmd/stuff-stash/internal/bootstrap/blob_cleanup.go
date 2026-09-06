package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	mediaapp "github.com/stuffstash/stuff-stash/internal/app/media"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func startBlobDeletionRechecks(ctx context.Context, repositories repositories, observer ports.Observer, cfg config.ThumbnailConfig) (func(), error) {
	worker, err := mediaapp.NewCleanupWorker(repositories.blobDeletionRechecks, repositories.blobs, ports.SystemClock{}, idgen.NewULIDGenerator(), observer, cfg.CleanupRecheckInterval, cfg.LeaseDuration, cfg.ProcessingTimeout)
	if err != nil {
		return nil, err
	}
	// Cleanup stays active when image generation is disabled and never decodes images.
	cfg.WorkerEnabled = true
	cfg.Concurrency = 1
	return startThumbnailWorkers(ctx, worker, observer, cfg), nil
}
