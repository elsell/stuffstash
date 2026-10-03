//go:build linux

package printworker_test

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/printstate"
	"github.com/stuffstash/stuff-stash/cli/internal/app/printworker"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type events chan string

func (e events) Event(_ context.Context, name string) {
	select {
	case e <- name:
	default:
	}
}

type timers struct{}

func (timers) Wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (timers) Delay(int) time.Duration { return time.Millisecond }

type registry struct {
	mu                  sync.Mutex
	assigned            []printing.RegisteredPrinter
	heartbeats, reports int
	changed             chan struct{}
}

func (r *registry) notify() {
	select {
	case r.changed <- struct{}{}:
	default:
	}
}
func (r *registry) Heartbeat(context.Context, string) error {
	r.mu.Lock()
	r.heartbeats++
	r.mu.Unlock()
	r.notify()
	return nil
}
func (r *registry) Printers(context.Context) ([]printing.RegisteredPrinter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]printing.RegisteredPrinter(nil), r.assigned...), nil
}
func (r *registry) Report(context.Context, string, printing.Readiness) error {
	r.mu.Lock()
	r.reports++
	r.mu.Unlock()
	r.notify()
	return nil
}

type devices struct {
	mu      sync.Mutex
	active  map[string]bool
	opened  map[string]int
	offline string
	changed chan struct{}
}

func (d *devices) Open(_ context.Context, p printing.RegisteredPrinter) (ports.PrinterConnection, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.opened[p.DeviceID]++
	select {
	case d.changed <- struct{}{}:
	default:
	}
	if p.DeviceID == d.offline {
		return nil, errors.New("USB disconnected")
	}
	if d.active[p.DeviceID] {
		return nil, ports.ErrDeviceInUse
	}
	d.active[p.DeviceID] = true
	return &blockedDevice{devices: d, id: p.DeviceID}, nil
}

type blockedDevice struct {
	devices *devices
	id      string
}

func (p *blockedDevice) Readiness(ctx context.Context) (printing.Readiness, error) {
	<-ctx.Done()
	return printing.Readiness{}, ctx.Err()
}
func (*blockedDevice) Submit(context.Context, printing.Label) (printing.Submission, error) {
	return printing.Submission{}, errors.New("not ready")
}
func (*blockedDevice) Observe(context.Context, printing.Submission) (printing.Observation, error) {
	return printing.Observation{}, errors.New("no submission")
}
func (p *blockedDevice) Close() error {
	p.devices.mu.Lock()
	defer p.devices.mu.Unlock()
	delete(p.devices.active, p.id)
	return nil
}

func runtimeFixture(t *testing.T) (*printworker.Runtime, *registry, *devices) {
	t.Helper()
	r := &registry{changed: make(chan struct{}, 1)}
	d := &devices{active: map[string]bool{}, opened: map[string]int{}, changed: make(chan struct{}, 1)}
	worker, _, _, _ := fixture(t)
	worker.Waiter = timers{}
	return &printworker.Runtime{Registry: r, Devices: d, State: printstate.Store{Directory: filepath.Join(t.TempDir(), "private")}, Worker: *worker, Backoff: timers{}, HeartbeatInterval: time.Millisecond, PollInterval: time.Millisecond}, r, d
}

func TestBlockedAndOfflineDevicesDoNotBlockConnectorHeartbeatOrEachOther(t *testing.T) {
	runtime, r, d := runtimeFixture(t)
	r.assigned = []printing.RegisteredPrinter{{ID: "first", DeviceID: "usb-one"}, {ID: "second", DeviceID: "usb-two"}}
	d.offline = "usb-two"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runtime.Run(ctx) }()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for {
		r.mu.Lock()
		beats := r.heartbeats
		r.mu.Unlock()
		d.mu.Lock()
		both := d.opened["usb-one"] > 0 && d.opened["usb-two"] > 0
		d.mu.Unlock()
		if beats >= 2 && both {
			break
		}
		select {
		case <-r.changed:
		case <-d.changed:
		case <-deadline.C:
			t.Fatal("one printer blocked the connector")
		}
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("shutdown: %v", err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.active) != 0 {
		t.Fatal("worker returned before releasing device locks")
	}
}
func TestDuplicatePhysicalBindingsNeverOpenCompetingQueues(t *testing.T) {
	runtime, r, d := runtimeFixture(t)
	r.assigned = []printing.RegisteredPrinter{{ID: "one", DeviceID: "same-device"}, {ID: "two", DeviceID: "same-device"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runtime.Run(ctx) }()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for {
		r.mu.Lock()
		reports := r.reports
		r.mu.Unlock()
		if reports >= 2 {
			break
		}
		select {
		case <-r.changed:
		case <-deadline.C:
			t.Fatal("duplicate bindings not reported")
		}
	}
	cancel()
	<-done
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.opened) != 0 {
		t.Fatal("duplicate physical device was opened")
	}
}

func TestOfflinePrinterReconcilesCompletedJournalBeforeAnyDeviceAccess(t *testing.T) {
	for _, retired := range []bool{false, true} {
		t.Run(strconv.FormatBool(retired), func(t *testing.T) {
			runtime, r, d := runtimeFixture(t)
			w, a, j, p := fixture(t)
			a.dropOutcome = true
			w.Config.Binding = "binding:printer"
			if err := w.Step(context.Background(), j, p); err == nil {
				t.Fatal("expected lost response")
			}
			// Model lease expiry after the server never acknowledged the completed outcome.
			a.status.Phase = printing.RemoteUncertain
			a.status.Revision++
			recorded := make(events, 10)
			w.Observer = recorded
			locked, err := runtime.State.Acquire(context.Background(), "offline-device")
			if err != nil {
				t.Fatal(err)
			}
			if err = locked.Save(context.Background(), j.record); err != nil {
				t.Fatal(err)
			}
			locked.Close()
			w.Config.Binding = "binding"
			w.Waiter = timers{}
			runtime.Worker = *w
			r.assigned = []printing.RegisteredPrinter{{ID: "printer", DeviceID: "offline-device", Media: w.Config.Media, Retired: retired}}
			d.offline = "offline-device"
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- runtime.Run(ctx) }()
			deadline := time.NewTimer(time.Second)
			defer deadline.Stop()
		recovery:
			for {
				select {
				case event := <-recorded:
					if event == "cli.print.recovered" {
						break recovery
					}
				case <-deadline.C:
					t.Fatal("retired recovery did not reconcile")
				}
			}
			cancel()
			<-done
			if a.evidence.Outcome != printing.Completed || a.status.Phase != printing.RemoteCompleted {
				t.Fatal("durable completion was not reconciled")
			}
			if a.claims != 1 || p.submissions != 2 {
				t.Fatal("retirement created another print attempt")
			}
			d.mu.Lock()
			defer d.mu.Unlock()
			if retired && len(d.opened) != 0 {
				t.Fatal("retired recovery required physical printer")
			}
			locked, err = runtime.State.Acquire(context.Background(), "offline-device")
			if err != nil {
				t.Fatal(err)
			}
			defer locked.Close()
			record, err := locked.Load(context.Background())
			if err != nil || record != nil {
				t.Fatal("settled journal not cleared")
			}
		})
	}
}
