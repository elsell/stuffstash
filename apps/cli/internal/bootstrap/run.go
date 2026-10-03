package bootstrap

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/oidcauth"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"github.com/stuffstash/stuff-stash/cli/internal/version"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }
func Run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	options, err := app.Parse(args, getenv)
	output := presentation.Output{Stdout: stdout, Stderr: stderr, JSON: options.JSON}
	if err == nil && options.Help {
		_, err = io.WriteString(stdout, Help)
		return exit(output, err)
	}
	if err == nil && len(options.Command) == 1 && options.Command[0] == "version" {
		return exit(output, output.Result(version.Current()))
	}
	if err == nil && len(options.Command) == 2 && options.Command[0] == "printers" && (options.Command[1] == "catalog" || options.Command[1] == "discover") {
		return exit(output, printerCommand(ctx, options.Command[1], getenv, output))
	}
	if err == nil {
		err = oidcauth.ValidateURL(options.Server, options.AllowLoopbackHTTP)
	}
	if err != nil {
		return exit(output, err)
	}
	options.Server = oidcauth.CanonicalServer(options.Server)
	if len(options.Command) == 3 && options.Command[0] == "connectors" && options.Command[1] == "print" && options.Command[2] == "register" {
		return exit(output, registerPrintConnector(ctx, options, getenv, output))
	}
	if len(options.Command) == 3 && options.Command[0] == "connectors" && options.Command[1] == "print" && options.Command[2] == "run" {
		return exit(output, runPrintConnector(ctx, options, getenv, output))
	}
	var store ports.Credentials = credentials.Keyring{}
	if options.CredentialFile != "" {
		store = credentials.File{Path: options.CredentialFile}
	}
	client := &http.Client{Timeout: 30 * time.Second}
	clock := systemClock{}
	runner := app.Runner{API: func(server, token string) (ports.API, error) { return httpapi.New(server, token, client) }, Auth: oidcauth.Adapter{HTTP: client, Clock: clock, Output: output, Browser: oidcauth.SystemBrowser{}, AllowLoopbackHTTP: options.AllowLoopbackHTTP}, Credentials: store, Clock: clock, Output: output, Observer: presentation.SilentObserver{}}
	return exit(output, runner.Run(ctx, options))
}
func exit(output ports.Output, err error) int {
	if err == nil {
		return 0
	}
	category, message := "configuration", err.Error()
	var typed *ports.Error
	if errors.As(err, &typed) {
		category, message = typed.Category, typed.Message
	}
	if errors.Is(err, ports.ErrNotLoggedIn) {
		category = "authentication"
	}
	output.Error(category, message)
	if category == "usage" || category == "configuration" {
		return 2
	}
	return 1
}

const Help = `Stuff Stash CLI

  stuffstash login --server https://stash.example [--device-code]
  stuffstash logout --server https://stash.example
  stuffstash inventories list --tenant ID
  stuffstash assets list --tenant ID --inventory ID [--limit N --cursor CURSOR]
  stuffstash assets show ID
  stuffstash assets create --kind item|container|location --title TITLE
  stuffstash assets update ID --title TITLE
  stuffstash assets move ID --parent ID|root
  stuffstash assets archive ID
  stuffstash assets restore ID
  stuffstash version
  stuffstash printers discover
  stuffstash printers catalog [--json]
  stuffstash connectors print register --name NAME
  stuffstash connectors print run --connector ID [--journal-dir PATH]

Context: --server, --tenant, --inventory or STUFF_STASH_CLI_SERVER,
STUFF_STASH_CLI_TENANT, STUFF_STASH_CLI_INVENTORY. No implicit inventory selection.
Finite commands accept --json. Mutations accept --idempotency-key.
Headless credential storage: explicitly set STUFF_STASH_CLI_CREDENTIAL_FILE.
Local printer discovery and catalog export do not need login.
Connector secrets: use the OS store or explicitly set
STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE for headless hosts.
Registration prints a browser approval URL and short code.
The Linux USB worker consumes all assigned printers. Keep its journal directory persistent.
`
