package agentmodel

import (
	"context"
	"errors"
	"sync"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const MaxEvaluationWorkerConcurrency = ports.MaxEvaluationWorkerConcurrency

func (w EvaluationWorker) Drain(ctx context.Context, limit, concurrency int) error {
	if limit < 1 || limit > ports.MaxEvaluationRunPageLimit || concurrency < 1 || concurrency > MaxEvaluationWorkerConcurrency {
		return apperrors.ErrValidation
	}
	if w.deps.Runs == nil || w.deps.Clock == nil {
		return apperrors.ErrPrecondition
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	references, err := w.deps.Runs.RunnableEvaluationRuns(ctx, w.deps.Clock.Now(), limit)
	if err != nil {
		return err
	}
	if len(references) > limit {
		return apperrors.ErrPrecondition
	}
	jobs := make(chan ports.EvaluationRunReference, len(references))
	failures := make(chan error, len(references))
	for _, reference := range references {
		jobs <- reference
	}
	close(jobs)
	var running sync.WaitGroup
	for worker := 0; worker < min(concurrency, len(references)); worker++ {
		running.Add(1)
		go func() {
			defer running.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case reference, ok := <-jobs:
					if !ok {
						return
					}
					if err := w.Process(ctx, reference); err != nil {
						failures <- err
					}
				}
			}
		}()
	}
	running.Wait()
	close(failures)
	if err := ctx.Err(); err != nil {
		return err
	}
	var collected []error
	for err := range failures {
		collected = append(collected, err)
	}
	return errors.Join(collected...)
}
