package bootstrap

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/pairingcrypto"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func configurePrintConnectors(application app.App, repositories repositories, authorization ports.PrintingAuthorization, cfg config.PrintConnectorConfig) (app.App, error) {
	if authorization == nil || repositories.printConnectors == nil {
		return application, errors.New("print connector authorization or persistence unavailable")
	}
	policy := printregistry.ConnectorPolicy{PublicWebBaseURL: cfg.PublicWebBaseURL, PairingLifetime: cfg.PairingLifetime, CredentialLifetime: cfg.CredentialLifetime, ActivationLifetime: cfg.ActivationLifetime, AuthorizationTimeout: cfg.AuthorizationTimeout, ReportMaxAge: cfg.ReportMaxAge}
	return application.WithPrintConnectors(repositories.printConnectors, authorization, pairingcrypto.Secrets{}, policy), nil
}
func startPrintConnectorWorker(ctx context.Context, application app.App, observer ports.Observer, cfg config.PrintConnectorConfig) {
	go runPeriodicDrain(ctx, cfg.PollInterval, func() {
		if err := application.PrintConnectors().DrainAuthorization(ctx, cfg.BatchSize); err != nil {
			observer.Record(ctx, ports.Event{Name: ports.EventPrintConnectorSyncFailed, Message: "print connector authorization reconciliation failed"})
		}
	})
}
