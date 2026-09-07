package bootstrap

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/worklimit"
	mediaapp "github.com/stuffstash/stuff-stash/internal/app/media"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildThumbnailRuntime(repositories repositories, cfg config.ThumbnailConfig, observer ports.Observer) (ports.ThumbnailReader, *mediaapp.Worker, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	admission, err := worklimit.New(cfg.Concurrency)
	if err != nil {
		return nil, nil, err
	}
	processor, err := mediaapp.NewProcessor(repositories.attachments, repositories.blobs, repositories.imageBatch, repositories.thumbnailGuard, cfg.PublicationTimeout)
	if err != nil {
		return nil, nil, err
	}
	reader, err := mediaapp.NewReader(processor, admission, cfg.ForegroundCachePollInterval)
	if err != nil {
		return nil, nil, err
	}
	worker, err := mediaapp.NewWorker(repositories.thumbnailQueue, processor, admission, ports.SystemClock{}, idgen.NewULIDGenerator(), observer, mediaapp.WorkerConfig{
		MaxAttempts: cfg.MaxAttempts, LeaseDuration: cfg.LeaseDuration, ProcessingTimeout: cfg.ProcessingTimeout, RetryBase: cfg.RetryBase, RetryMax: cfg.RetryMax,
	})
	if err != nil {
		return nil, nil, err
	}
	return reader, worker, nil
}
