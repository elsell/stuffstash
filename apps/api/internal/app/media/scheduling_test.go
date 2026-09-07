package media

import (
	"context"
	"errors"
	"runtime"
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
