package printworker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Runtime struct {
	SoftwareReport                  *printing.ConnectorReport
	Registry                        ports.PrintRegistry
	Devices                         ports.DevicePrinters
	State                           ports.PrintState
	Worker                          Worker
	Backoff                         ports.PrintBackoff
	HeartbeatInterval, PollInterval time.Duration
}

type activePrinter struct {
	registration printing.RegisteredPrinter
	cancel       context.CancelFunc
	done         chan struct{}
	retired      *atomic.Bool
}

// Run keeps connector liveness independent of serialized physical-device work.
// It joins changed/removed workers before opening their replacement connection.
func (r *Runtime) Run(ctx context.Context) error {
	if r.HeartbeatInterval <= 0 || r.PollInterval <= 0 {
		return errors.New("invalid print runtime intervals")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	running := make(map[string]activePrinter)
	fatal := make(chan error, 1)
	var workers sync.WaitGroup
	defer func() { cancel(); workers.Wait() }()
	failures := 0
	for {
		select {
		case err := <-fatal:
			return err
		default:
		}
		err := r.Registry.Heartbeat(ctx, r.Worker.Config.SessionID, r.SoftwareReport)
		var registrations []printing.RegisteredPrinter
		if err == nil {
			registrations, err = r.Registry.Printers(ctx)
		}
		if terminal(err) {
			return err
		}
		delay := r.HeartbeatInterval
		if err != nil {
			failures++
			delay = r.Backoff.Delay(failures)
			r.Worker.Observer.Event(ctx, "cli.print.connector_unavailable")
		} else {
			failures = 0
			desired := uniqueDevices(registrations)
			for id, active := range running {
				next, exists := desired[id]
				previous := active.registration
				previous.Retired = next.Retired
				if exists && previous == next {
					active.retired.Store(next.Retired)
					active.registration = next
					running[id] = active
				}
				if !exists || next != previous {
					active.cancel()
					<-active.done
					delete(running, id)
				}
			}
			for id, registration := range desired {
				if _, exists := running[id]; exists {
					continue
				}
				child, stop := context.WithCancel(ctx)
				active := activePrinter{registration: registration, cancel: stop, done: make(chan struct{}), retired: &atomic.Bool{}}
				active.retired.Store(registration.Retired)
				running[id] = active
				workers.Add(1)
				go func() {
					defer workers.Done()
					defer close(active.done)
					if err := r.runPrinter(child, registration, active.retired); terminal(err) {
						select {
						case fatal <- err:
							cancel()
						default:
						}
					}
				}()
			}
			for _, registration := range registrations {
				if !registration.Retired {
					if _, exists := desired[registration.ID]; !exists {
						if err := r.Registry.Report(ctx, registration.ID, printing.Readiness{State: printing.NeedsAttention, Reason: printing.ActiveSubmission}); terminal(err) {
							return err
						}
					}
				}
			}
		}
		if err = r.Worker.Waiter.Wait(ctx, delay); err != nil {
			select {
			case failure := <-fatal:
				return failure
			default:
				return err
			}
		}
	}
}

func uniqueDevices(registrations []printing.RegisteredPrinter) map[string]printing.RegisteredPrinter {
	counts := make(map[string]int)
	for _, registration := range registrations {
		counts[registration.DeviceID]++
	}
	desired := make(map[string]printing.RegisteredPrinter)
	for _, registration := range registrations {
		if registration.DeviceID != "" && counts[registration.DeviceID] == 1 {
			desired[registration.ID] = registration
		}
	}
	return desired
}

func terminal(err error) bool {
	var failure *ports.Error
	return errors.As(err, &failure) && (failure.Category == "authentication" || failure.Category == "authorization")
}
