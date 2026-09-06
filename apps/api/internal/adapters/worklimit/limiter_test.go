package worklimit

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCapacityIsSharedBetweenForegroundAndBackground(t *testing.T) {
	limiter, err := New(2)
	if err != nil {
		t.Fatal(err)
	}
	first, err := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	second, err := limiter.Acquire(context.Background(), ports.ImageWorkForeground)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	waiting := acquireAsync(limiter, ctx, ports.ImageWorkBackground)
	awaitWaiters(t, limiter, 0, 1)
	select {
	case <-waiting:
		t.Fatal("shared capacity exceeded")
	default:
	}
	first()
	third := awaitPermit(t, waiting)
	// Double release must not admit a fourth image while second and third run.
	first()
	fourth := acquireAsync(limiter, ctx, ports.ImageWorkForeground)
	awaitWaiters(t, limiter, 1, 0)
	select {
	case <-fourth:
		t.Fatal("double release increased capacity")
	default:
	}
	second()
	awaitPermit(t, fourth)()
	third()
}

func TestWaitingForegroundHasPriorityAndEachClassIsFIFO(t *testing.T) {
	limiter, _ := New(1)
	active, _ := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	background := acquireAsync(limiter, ctx, ports.ImageWorkBackground)
	awaitWaiters(t, limiter, 0, 1)
	foregroundOne := acquireAsync(limiter, ctx, ports.ImageWorkForeground)
	awaitWaiters(t, limiter, 1, 1)
	foregroundTwo := acquireAsync(limiter, ctx, ports.ImageWorkForeground)
	awaitWaiters(t, limiter, 2, 1)
	active()
	first := awaitPermit(t, foregroundOne)
	select {
	case <-background:
		t.Fatal("background bypassed foreground")
	default:
	}
	first()
	awaitPermit(t, foregroundTwo)()
	awaitPermit(t, background)()
}

func TestCancellationDoesNotLeakCapacity(t *testing.T) {
	limiter, _ := New(1)
	active, _ := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		release, err := limiter.Acquire(ctx, ports.ImageWorkForeground)
		if release != nil {
			release()
		}
		result <- err
	}()
	awaitWaiters(t, limiter, 1, 0)
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled waiter did not stop")
	}
	active()
	release, err := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestCancelledContextCannotTakeFreeCapacity(t *testing.T) {
	limiter, _ := New(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	release, err := limiter.Acquire(ctx, ports.ImageWorkForeground)
	if release != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled work admitted")
	}
}

func TestInvalidCapacityAndPriorityAreRejected(t *testing.T) {
	for _, capacity := range []int{-1, 0} {
		if _, err := New(capacity); err == nil {
			t.Fatal("invalid capacity accepted")
		}
	}
	limiter, _ := New(1)
	if release, err := limiter.Acquire(context.Background(), ports.ImageWorkPriority("unknown")); err == nil || release != nil {
		t.Fatal("unknown work priority accepted")
	}
}

func acquireAsync(limiter *Limiter, ctx context.Context, priority ports.ImageWorkPriority) <-chan func() {
	result := make(chan func(), 1)
	go func() {
		release, err := limiter.Acquire(ctx, priority)
		if err == nil {
			result <- release
		}
	}()
	return result
}

func awaitPermit(t *testing.T, result <-chan func()) func() {
	t.Helper()
	select {
	case release := <-result:
		return release
	case <-time.After(time.Second):
		t.Fatal("work never received available capacity")
		return nil
	}
}

// Inspect the queue only to synchronize contenders, never to assert scheduling outcomes.
func awaitWaiters(t *testing.T, limiter *Limiter, foreground, background int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		limiter.mu.Lock()
		ready := len(limiter.foreground) == foreground && len(limiter.background) == background
		limiter.mu.Unlock()
		if ready {
			return
		}
		runtime.Gosched()
	}
	t.Fatal("contenders did not reach admission queue")
}
