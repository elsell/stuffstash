package bootstrap

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type timerWaiter struct{}

func (timerWaiter) Wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func connectorCredentialStore(options app.Options) ports.ConnectorCredentials {
	if options.ConnectorCredentialFile != "" {
		return credentials.ConnectorFile{Path: options.ConnectorCredentialFile}
	}
	return credentials.ConnectorKeyring{}
}
func registerPrintConnector(ctx context.Context, options app.Options, getenv func(string) string, output ports.Output) error {
	var candidates []ports.PairingCandidate
	var err error
	if options.Command[2] != "rotate" {
		candidates, err = discoverPairingCandidates(ctx, getenv)
		if err != nil {
			return err
		}
	}
	api, err := httpapi.NewPairing(options.Server, &http.Client{Timeout: 30 * time.Second})
	if err != nil {
		return err
	}
	receipts := presentation.NewProtocolReceipts(output)
	api.Receipts = receipts
	defer func() {
		drainCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		receipts.Close(drainCtx)
	}()
	interval := 5 * time.Second
	if raw := getenv("STUFF_STASH_CLI_PAIRING_POLL_INTERVAL"); raw != "" {
		interval, err = time.ParseDuration(raw)
		if err != nil || interval < time.Second || interval > time.Minute {
			return ports.Failure("configuration", "Set STUFF_STASH_CLI_PAIRING_POLL_INTERVAL to a duration from 1s through 1m.")
		}
	}
	registrar := app.ConnectorRegistrar{Receipts: receipts, ReceiptDrainTimeout: time.Second, API: api, Credentials: connectorCredentialStore(options), Keys: pairingkeys.Keys{}, Clock: systemClock{}, Waiter: timerWaiter{}, Output: output, PollInterval: interval}
	if options.Command[2] == "rotate" {
		return registrar.Rotate(ctx, options.Server, options.ConnectorID)
	}
	return registrar.Register(ctx, options.Server, options.ConnectorName, candidates)
}

func discoverPairingCandidates(ctx context.Context, getenv func(string) string) ([]ports.PairingCandidate, error) {
	candidates := []ports.PairingCandidate{}
	for _, runtime := range BuiltinPrinters(getenv) {
		devices, err := runtime.Discovery.Discover(ctx)
		if err != nil {
			return nil, err
		}
		for _, device := range devices {
			candidates = append(candidates, ports.PairingCandidate{ID: "candidate-" + strconv.Itoa(len(candidates)+1), Name: device.Model, AdapterID: runtime.Printer.Descriptor().ID, DeviceID: device.ID})
		}
	}
	return candidates, nil
}
