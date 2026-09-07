package observability

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProfilerPushesProfilesWithoutPublicListener(t *testing.T) {
	var uploads atomic.Int64
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "stack" || password != "test-password" {
			t.Error("profiling credentials missing")
		}
		count, err := io.Copy(io.Discard, r.Body)
		if err != nil || count == 0 {
			t.Error("empty profile")
		}
		uploads.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()
	session, err := NewProfiler(context.Background(), profileTestConfig(collector.URL), discardObserver{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := session.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if uploads.Load() == 0 {
		t.Fatal("no profiles uploaded")
	}
}

func TestProfilerRejectsRedirectsAndSanitizesSDKDiagnostics(t *testing.T) {
	var redirected atomic.Int64
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected.Add(1) }))
	defer destination.Close()
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", destination.URL+"/private-secret")
		http.Error(w, "private-secret", http.StatusTemporaryRedirect)
	}))
	defer collector.Close()
	observer := &profileEvents{}
	session, err := NewProfiler(context.Background(), profileTestConfig(collector.URL), observer)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := session.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if redirected.Load() != 0 {
		t.Fatal("profiler followed a credential-bearing redirect")
	}
	observer.mu.Lock()
	defer observer.mu.Unlock()
	if len(observer.events) == 0 {
		t.Fatal("profile delivery failure was not observed")
	}
	for _, event := range observer.events {
		if strings.Contains(event.Message, "private-secret") {
			t.Fatal("raw profiler diagnostic leaked")
		}
		for _, value := range event.Fields {
			if strings.Contains(value, "private-secret") {
				t.Fatal("profile field leaked")
			}
		}
	}
}

func TestDisabledProfilerStartsNoCollector(t *testing.T) {
	session, err := NewProfiler(context.Background(), config.ProfilingConfig{}, discardObserver{})
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func profileTestConfig(endpoint string) config.ProfilingConfig {
	return config.ProfilingConfig{Enabled: true, Endpoint: endpoint, Username: "stack", Password: "test-password", ServiceName: "test-api", UploadInterval: time.Hour, RequestTimeout: time.Second}
}

type profileEvents struct {
	mu     sync.Mutex
	events []ports.Event
}

func (p *profileEvents) Record(_ context.Context, event ports.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, event)
}

func TestProfilerOwnsSamplingUntilConcurrentStopsComplete(t *testing.T) {
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.Copy(io.Discard, r.Body) }))
	defer collector.Close()
	previous := runtime.SetMutexProfileFraction(7)
	defer runtime.SetMutexProfileFraction(previous)
	cfg := profileTestConfig(collector.URL)
	cfg.MutexFraction = 3
	session, err := NewProfiler(context.Background(), cfg, discardObserver{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Stop(context.Background())
	if got := runtime.SetMutexProfileFraction(-1); got != 3 {
		t.Fatalf("active fraction = %d", got)
	}
	if _, err := NewProfiler(context.Background(), cfg, discardObserver{}); err == nil {
		t.Fatal("duplicate profiler accepted")
	}
	var workers sync.WaitGroup
	for range 3 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := session.Stop(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	if got := runtime.SetMutexProfileFraction(-1); got != 7 {
		t.Fatalf("restored fraction = %d", got)
	}
	next, err := NewProfiler(context.Background(), cfg, discardObserver{})
	if err != nil {
		t.Fatal(err)
	}
	if err := next.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
