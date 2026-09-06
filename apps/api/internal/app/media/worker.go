package media

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type WorkerConfig struct {
	MaxAttempts                                           int
	LeaseDuration, ProcessingTimeout, RetryBase, RetryMax time.Duration
}

func (c WorkerConfig) Validate() error {
	if c.MaxAttempts < 1 || c.MaxAttempts > 100 || c.ProcessingTimeout <= 0 || c.LeaseDuration <= c.ProcessingTimeout || c.RetryBase <= 0 || c.RetryMax < c.RetryBase {
		return errors.New("invalid thumbnail worker configuration")
	}
	return nil
}

type Worker struct {
	queue     ports.ThumbnailJobQueue
	processor ports.ThumbnailJobProcessor
	admission ports.ImageWorkAdmission
	clock     ports.Clock
	ids       ports.IDGenerator
	observer  ports.Observer
	config    WorkerConfig
}

func NewWorker(queue ports.ThumbnailJobQueue, processor ports.ThumbnailJobProcessor, admission ports.ImageWorkAdmission, clock ports.Clock, ids ports.IDGenerator, observer ports.Observer, config WorkerConfig) (*Worker, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if queue == nil || processor == nil || admission == nil || clock == nil || ids == nil || observer == nil {
		return nil, errors.New("thumbnail worker dependencies are required")
	}
	return &Worker{queue: queue, processor: processor, admission: admission, clock: clock, ids: ids, observer: observer, config: config}, nil
}

// Drain handles one image. Runtime scheduling and shutdown ownership belong to bootstrap.
func (w *Worker) Drain(ctx context.Context) (bool, error) {
	release, err := w.admission.Acquire(ctx, ports.ImageWorkBackground)
	if err != nil {
		return false, err
	}
	defer release()
	now := w.clock.Now()
	claims, err := w.queue.ClaimThumbnailJobs(ctx, w.ids.NewID(), 1, now, now.Add(w.config.LeaseDuration))
	if err != nil || len(claims) == 0 {
		return false, err
	}
	if len(claims) != 1 {
		return false, errors.New("thumbnail queue exceeded claim limit")
	}
	claim := claims[0]
	if claim.Attempts > w.config.MaxAttempts {
		return true, w.resolve(ctx, claim, ports.ThumbnailJobResolution{Status: ports.ThumbnailJobFailed, At: w.clock.Now(), Failure: ports.ThumbnailFailureProcessing})
	}
	remaining := claim.ClaimedUntil.Sub(w.clock.Now())
	if remaining <= 0 {
		return true, ports.ErrOutboxClaimLost
	}
	budget := remaining - (w.config.LeaseDuration - w.config.ProcessingTimeout)
	var processErr error
	if budget <= 0 {
		processErr = ports.ErrOutboxClaimLost
	} else {
		if budget > w.config.ProcessingTimeout {
			budget = w.config.ProcessingTimeout
		}
		processing, cancel := context.WithTimeout(ctx, budget)
		processErr = w.processor.ProcessThumbnailJob(processing, claim)
		// A late nil result must not acknowledge work after its processing deadline.
		if processErr == nil {
			processErr = processing.Err()
		}
		cancel()
	}

	if err := ctx.Err(); err != nil {
		return true, err
	}
	resolution := ports.ThumbnailJobResolution{Status: ports.ThumbnailJobCompleted, At: w.clock.Now()}
	if processErr != nil {
		resolution.Failure = ports.ThumbnailFailureProcessing
		if claim.Attempts >= w.config.MaxAttempts {
			resolution.Status = ports.ThumbnailJobFailed
		} else {
			resolution.Status = ports.ThumbnailJobPending
			resolution.NextAttemptAt = resolution.At.Add(w.retryDelay(claim.Attempts))
		}
	}
	return true, w.resolve(ctx, claim, resolution)
}

func (w *Worker) resolve(ctx context.Context, claim ports.ClaimedThumbnailJob, resolution ports.ThumbnailJobResolution) error {
	if err := w.queue.ResolveThumbnailJob(ctx, claim, resolution); err != nil {
		return err
	}
	w.observer.Record(ctx, ports.Event{Name: ports.EventThumbnailJobResolved, Message: "thumbnail work resolved", Fields: map[string]string{"outcome": string(resolution.Status), "reason": string(resolution.Failure)}})
	return nil
}

func (w *Worker) retryDelay(attempt int) time.Duration {
	delay := w.config.RetryBase
	for i := 1; i < attempt; i++ {
		if delay >= w.config.RetryMax/2 {
			return w.config.RetryMax
		}
		delay *= 2
	}
	return delay
}
