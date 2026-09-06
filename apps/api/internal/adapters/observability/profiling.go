package observability

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/grafana/pyroscope-go"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Runtime contention sampling is process-wide, so only one adapter may own it.
var profilerActive atomic.Bool

type Profiler struct {
	stopOnce      sync.Once
	done          chan struct{}
	session       *pyroscope.Profiler
	previousMutex int
}

func NewProfiler(ctx context.Context, cfg config.ProfilingConfig, observer ports.Observer) (*Profiler, error) {
	p := &Profiler{done: make(chan struct{})}
	if !cfg.Enabled {
		close(p.done)
		return p, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !profilerActive.CompareAndSwap(false, true) {
		return nil, errors.New("profiling already active")
	}
	p.previousMutex = runtime.SetMutexProfileFraction(cfg.MutexFraction)
	runtime.SetBlockProfileRate(cfg.BlockRate)
	session, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: cfg.ServiceName,
		Tags:            map[string]string{"service_version": cfg.ServiceVersion, "environment": cfg.Environment},
		ServerAddress:   cfg.Endpoint, BasicAuthUser: cfg.Username, BasicAuthPassword: cfg.Password,
		UploadRate: cfg.UploadInterval, DisableGCRuns: true,
		Logger:       profileLogger{ctx: context.WithoutCancel(ctx), observer: observer},
		HTTPClient:   &http.Client{Timeout: cfg.RequestTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		ProfileTypes: []pyroscope.ProfileType{pyroscope.ProfileCPU, pyroscope.ProfileInuseObjects, pyroscope.ProfileAllocObjects, pyroscope.ProfileInuseSpace, pyroscope.ProfileAllocSpace, pyroscope.ProfileGoroutines, pyroscope.ProfileMutexCount, pyroscope.ProfileMutexDuration, pyroscope.ProfileBlockCount, pyroscope.ProfileBlockDuration},
	})
	if err != nil {
		p.releaseSampling()
		return nil, errors.New("profiling startup failed")
	}
	p.session = session
	return p, nil
}

func (p *Profiler) Stop(ctx context.Context) error {
	if p.session == nil {
		return nil
	}
	p.stopOnce.Do(func() {
		go func() {
			defer close(p.done)
			defer p.releaseSampling()
			p.session.Flush(true)
			_ = p.session.Stop()
		}()
	})
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return errors.New("profiling shutdown timed out")
	}
}

func (p *Profiler) releaseSampling() {
	runtime.SetMutexProfileFraction(p.previousMutex)
	runtime.SetBlockProfileRate(0)
	profilerActive.Store(false)
}

type profileLogger struct {
	ctx      context.Context
	observer ports.Observer
}

func (profileLogger) Infof(string, ...interface{})  {}
func (profileLogger) Debugf(string, ...interface{}) {}
func (p profileLogger) Errorf(string, ...interface{}) {
	if p.observer != nil {
		p.observer.Record(p.ctx, ports.Event{Name: ports.EventProfilingDeliveryFailed, Message: "profile delivery failed"})
	}
}
