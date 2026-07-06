package bootstrap

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestVacuumImportJobCredentialsWorkerDoesNotDuplicateFailureEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	observer := &bootstrapRecordingObserver{cancelOn: ports.EventImportJobCredentialVacuumFailed, cancel: cancel}
	application := app.New(app.Dependencies{
		Observer:          observer,
		ImportSourceVault: bootstrapFailingImportSourceVault{},
	})

	vacuumImportJobCredentials(ctx, application, observer, time.Hour)

	events := observer.eventsNamed(ports.EventImportJobCredentialVacuumFailed)
	if len(events) != 1 {
		t.Fatalf("expected one import credential vacuum failure event, got %+v", observer.events)
	}
	event := events[0]
	if event.Fields["error_class"] != "credential_vacuum_failed" {
		t.Fatalf("expected app-level safe error class, got %+v", event.Fields)
	}
	if _, ok := event.Fields["error"]; ok {
		t.Fatalf("raw error field must not be recorded: %+v", event.Fields)
	}
	for _, value := range event.Fields {
		for _, unsafe := range []string{"secret", "password", "ciphertext", "/tmp/provider-key"} {
			if strings.Contains(value, unsafe) {
				t.Fatalf("credential vacuum worker event leaked unsafe value %q in %+v", unsafe, event.Fields)
			}
		}
	}
}

type bootstrapRecordingObserver struct {
	events   []ports.Event
	cancelOn ports.EventName
	cancel   context.CancelFunc
}

func (f *bootstrapRecordingObserver) Record(_ context.Context, event ports.Event) {
	f.events = append(f.events, event)
	if event.Name == f.cancelOn && f.cancel != nil {
		f.cancel()
	}
}

func (f *bootstrapRecordingObserver) eventsNamed(name ports.EventName) []ports.Event {
	var events []ports.Event
	for _, event := range f.events {
		if event.Name == name {
			events = append(events, event)
		}
	}
	return events
}

type bootstrapFailingImportSourceVault struct{}

func (bootstrapFailingImportSourceVault) StoreImportJobSource(context.Context, ports.ImportJobSourceScope, ports.ImportSourceRequest, time.Time, time.Time) error {
	return nil
}

func (bootstrapFailingImportSourceVault) ImportJobSourceRequest(context.Context, ports.ImportJobSourceScope) (ports.ImportSourceRequest, bool, error) {
	return ports.ImportSourceRequest{}, false, nil
}

func (bootstrapFailingImportSourceVault) DeleteImportJobSource(context.Context, ports.ImportJobSourceScope) (bool, error) {
	return false, nil
}

func (bootstrapFailingImportSourceVault) VacuumImportJobSources(context.Context, time.Time) ([]ports.ImportJobSourceScope, error) {
	return nil, errors.New("password=secret ciphertext=abc /tmp/provider-key")
}
