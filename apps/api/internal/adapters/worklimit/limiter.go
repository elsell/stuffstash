// Package worklimit provides in-process admission for foreground and background images.
package worklimit

import (
	"context"
	"errors"
	"sync"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type waiter struct {
	ready   chan struct{}
	granted bool
}

type Limiter struct {
	mu                     sync.Mutex
	capacity, active       int
	foreground, background []*waiter
}

var _ ports.ImageWorkAdmission = (*Limiter)(nil)

func (l *Limiter) ForegroundWaiting() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.foreground) > 0
}

func New(capacity int) (*Limiter, error) {
	if capacity <= 0 {
		return nil, errors.New("image work capacity must be positive")
	}
	return &Limiter{capacity: capacity}, nil
}

func (l *Limiter) Acquire(ctx context.Context, priority ports.ImageWorkPriority) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if priority != ports.ImageWorkForeground && priority != ports.ImageWorkBackground {
		return nil, errors.New("invalid image work priority")
	}
	waiting := &waiter{ready: make(chan struct{})}
	l.mu.Lock()
	if priority == ports.ImageWorkForeground {
		l.foreground = append(l.foreground, waiting)
	} else {
		l.background = append(l.background, waiting)
	}
	l.dispatch()
	l.mu.Unlock()
	select {
	case <-waiting.ready:
		if err := ctx.Err(); err != nil {
			l.release()
			return nil, err
		}
		var once sync.Once
		return func() { once.Do(l.release) }, nil
	case <-ctx.Done():
		l.mu.Lock()
		if waiting.granted {
			l.active--
		} else if priority == ports.ImageWorkForeground {
			l.foreground = removeWaiter(l.foreground, waiting)
		} else {
			l.background = removeWaiter(l.background, waiting)
		}
		l.dispatch()
		l.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (l *Limiter) release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.active--
	l.dispatch()
}

// dispatch runs under mu. Priority applies at admission, without preempting work.
func (l *Limiter) dispatch() {
	for l.active < l.capacity {
		queue := &l.foreground
		if len(*queue) == 0 {
			queue = &l.background
		}
		if len(*queue) == 0 {
			return
		}
		next := (*queue)[0]
		(*queue)[0] = nil
		*queue = (*queue)[1:]
		next.granted = true
		l.active++
		close(next.ready)
	}
}

func removeWaiter(queue []*waiter, target *waiter) []*waiter {
	for i, waiting := range queue {
		if waiting == target {
			copy(queue[i:], queue[i+1:])
			queue[len(queue)-1] = nil
			return queue[:len(queue)-1]
		}
	}
	return queue
}
