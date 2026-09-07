package media

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/worklimit"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestBackgroundYieldsAfterSmallAndResumesOnlyMissingVariants(t *testing.T) {
	p, _, _, guard, claim := processorFixture(t)
	limiter, _ := worklimit.New(1)
	p.admission = limiter
	release, _ := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	defer release()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	foreground := make(chan error, 1)
	go func() {
		done, err := limiter.Acquire(ctx, ports.ImageWorkForeground)
		if done != nil {
			done()
		}
		foreground <- err
	}()
	for !limiter.ForegroundWaiting() {
		if ctx.Err() != nil {
			t.Fatal("foreground never queued")
		}
		runtime.Gosched()
	}
	if err := p.ProcessThumbnailJob(ctx, claim); !errors.Is(err, ports.ErrThumbnailYielded) {
		t.Fatal("background did not yield", err)
	}
	if guard.writes != 1 {
		t.Fatal("generated beyond small checkpoint", guard.writes)
	}
	release()
	if err := <-foreground; err != nil {
		t.Fatal(err)
	}
	resume, err := limiter.Acquire(ctx, ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	defer resume()
	if err := p.ProcessThumbnailJob(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if guard.writes != 3 {
		t.Fatal("resume regenerated persisted small", guard.writes)
	}
}

func TestReadersAndWorkersDoNotDuplicateAnOwnedPhoto(t *testing.T) {
	p, source, _, guard, claim := processorFixture(t)
	limiter, _ := worklimit.New(2)
	p.admission = limiter
	release, _, owned := p.flights.TryStart(source.attachment.StorageKey)
	if !owned {
		t.Fatal("first owner rejected")
	}
	defer release()
	reader, err := NewReader(p, limiter)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, _, err := reader.ReadThumbnail(ctx, source.attachment, domain.ThumbnailVariantSmall); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("reader bypassed existing flight", err)
	}
	if err := p.ProcessThumbnailJob(context.Background(), claim); !errors.Is(err, ports.ErrThumbnailYielded) {
		t.Fatal("worker blocked or duplicated owned photo", err)
	}
	if guard.writes != 0 {
		t.Fatal("duplicate publisher ran")
	}
	release()
	if _, cached, err := reader.ReadThumbnail(context.Background(), source.attachment, domain.ThumbnailVariantSmall); err != nil || cached {
		t.Fatal("owner release did not allow fallback", err)
	}
	if guard.writes != 1 {
		t.Fatal("fallback not published exactly once")
	}
}

func TestWaitingReaderRecoversWhenActiveOwnerCancels(t *testing.T) {
	p, source, _, guard, _ := processorFixture(t)
	limiter, _ := worklimit.New(2)
	p.admission = limiter
	batch := &cancelFirstBatch{ImageBatchProcessor: p.images, entered: make(chan struct{})}
	p.images = batch
	waiting := make(chan struct{})
	p.flights = readerObservedFlights{ThumbnailFlights: p.flights, waiting: waiting}
	reader, err := NewReader(p, limiter)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ownerCtx, cancelOwner := context.WithCancel(ctx)
	defer cancelOwner()
	owner := make(chan error, 1)
	go func() {
		_, _, err := reader.ReadThumbnail(ownerCtx, source.attachment, domain.ThumbnailVariantSmall)
		owner <- err
	}()
	select {
	case <-batch.entered:
	case <-ctx.Done():
		t.Fatal("owner never entered")
	}
	waiter := make(chan error, 1)
	go func() {
		_, _, err := reader.ReadThumbnail(ctx, source.attachment, domain.ThumbnailVariantSmall)
		waiter <- err
	}()
	select {
	case <-waiting:
	case <-ctx.Done():
		t.Fatal("second reader never waited")
	}
	cancelOwner()
	if err := <-owner; !errors.Is(err, context.Canceled) {
		t.Fatal("owner did not cancel", err)
	}
	if err := <-waiter; err != nil {
		t.Fatal("waiting reader did not recover", err)
	}
	if batch.calls.Load() != 2 || batch.maximum.Load() != 1 || guard.writes != 1 {
		t.Fatalf("unexpected generation: calls=%d maximum=%d writes=%d", batch.calls.Load(), batch.maximum.Load(), guard.writes)
	}
}

type cancelFirstBatch struct {
	ports.ImageBatchProcessor
	entered                chan struct{}
	calls, active, maximum atomic.Int32
}

func (b *cancelFirstBatch) CreateThumbnails(ctx context.Context, request ports.ImageDerivativesRequest, publish func(domain.ThumbnailVariant, ports.ImageDerivative) error) error {
	active := b.active.Add(1)
	defer b.active.Add(-1)
	for maximum := b.maximum.Load(); active > maximum; maximum = b.maximum.Load() {
		if b.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	if b.calls.Add(1) == 1 {
		close(b.entered)
		<-ctx.Done()
		return ctx.Err()
	}
	return b.ImageBatchProcessor.CreateThumbnails(ctx, request, publish)
}
