package bootstrap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/printprocess"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/printstate"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/app/printworker"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func runPrintConnector(ctx context.Context, options app.Options, getenv func(string) string, output ports.Output) error {
	if runtime.GOOS != "linux" {
		return ports.Failure("configuration", "USB print workers require Linux. Run the worker on Linux.")
	}
	if options.ConnectorID == "" {
		return ports.Failure("usage", "Select the registered connector with --connector.")
	}
	config, err := printConfig(options, getenv)
	if err != nil {
		return err
	}
	registration, err := connectorCredentialStore(options).Load(ctx, options.Server, options.ConnectorID)
	if err != nil {
		return err
	}
	clock := systemClock{}
	if !registration.ExpiresAt.After(clock.Now()) {
		return ports.Failure("authentication", "The connector credential expired. Pair this connector again.")
	}
	receipts := presentation.NewProtocolReceipts(output)
	defer func() {
		drain, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		receipts.Close(drain)
	}()
	api, err := httpapi.New(registration.Server, registration.Credential, &http.Client{Timeout: 30 * time.Second}, httpapi.Options{Receipts: receipts})
	if err != nil {
		return err
	}
	identities := printprocess.Identities{}
	session, err := identities.Attempt()
	if err != nil {
		return err
	}
	binding, _ := json.Marshal([]string{registration.Server, registration.TenantID, registration.InventoryID, registration.ConnectorID})
	fingerprint := sha256.Sum256(binding)
	worker := printworker.Worker{Receipts: receipts, Jobs: api, Clock: clock, Waiter: printprocess.Waiter{}, Identity: identities, Observer: presentation.SilentObserver{}, Config: printworker.Config{Binding: hex.EncodeToString(fingerprint[:]), SessionID: session, MaxArtifactBytes: config.artifactBytes, LeaseSafety: config.safety, ObserveInterval: config.observe, ReadinessTimeout: config.readiness}}
	service := printworker.Runtime{SoftwareReport: connectorSoftwareReport(getenv), Registry: api, Devices: registeredDevices{runtimes: BuiltinPrinters(getenv)}, State: printstate.Store{Directory: config.directory}, Worker: worker, Backoff: printprocess.Backoff{Minimum: config.minimum, Maximum: config.maximum}, HeartbeatInterval: config.heartbeat, PollInterval: config.poll}
	if err = output.Notice("Print connector running. Printer availability and jobs are visible in Stuff Stash. Press Ctrl-C to stop."); err != nil {
		return err
	}
	active, cancel := context.WithDeadline(ctx, registration.ExpiresAt)
	defer cancel()
	err = service.Run(active)
	if ctx.Err() != nil && errors.Is(err, context.Canceled) {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) && !registration.ExpiresAt.After(clock.Now()) {
		return ports.Failure("authentication", "The connector credential expired. Pair this connector again.")
	}
	return err
}
