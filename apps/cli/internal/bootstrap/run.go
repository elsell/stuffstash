package bootstrap

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/inputfiles"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/labelfiles"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/oidcauth"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/terminal"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"github.com/stuffstash/stuff-stash/cli/internal/version"
	"golang.org/x/term"
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
		return exit(output, output.Result(version.Current(supportsUSB(BuiltinPrinters(getenv), runtime.GOOS))))
	}
	if err == nil && len(options.Command) == 2 && options.Command[0] == "printers" && (options.Command[1] == "catalog" || options.Command[1] == "discover") {
		return exit(output, printerCommand(ctx, options.Command[1], getenv, output))
	}
	if err != nil {
		return exit(output, err)
	}
	var contextStore contexts.Store
	connector := len(options.Command) > 0 && options.Command[0] == "connectors"
	if !connector {
		local := len(options.Command) > 0 && options.Command[0] == "context"
		contextStore, err = configuredContexts(getenv, local || options.Server == "" || options.Selection.Context != "")
		if err != nil {
			return exit(output, err)
		}
		if local {
			return exit(output, (app.Runner{Contexts: contextStore, Output: output}).Run(ctx, options))
		}
		options.Server, err = contextServer(ctx, contextStore, options)
		if err != nil {
			return exit(output, err)
		}
	}

	if err == nil {
		err = oidcauth.ValidateURL(options.Server, options.AllowLoopbackHTTP)
	}
	if err != nil {
		return exit(output, err)
	}
	options.Server = oidcauth.CanonicalServer(options.Server)
	if len(options.Command) == 3 && options.Command[0] == "connectors" && options.Command[1] == "print" && (options.Command[2] == "register" || options.Command[2] == "rotate") {
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
	var picker ports.Selector
	var textInput ports.TextInput
	var secretInput ports.SecretInput
	terminalOutput, _ := stdout.(*os.File)
	terminalErrors, _ := stderr.(*os.File)
	if !options.NoInput && !options.JSON && getenv("TERM") != "dumb" && terminal.Available(os.Stdin, terminalOutput, terminalErrors) {
		prompt := terminal.Picker{Input: os.Stdin, Output: terminalErrors, Color: options.Color == "always" || (options.Color == "auto" && getenv("NO_COLOR") == "")}
		picker = prompt
		textInput = prompt
		secretInput = prompt
	}
	runner := app.Runner{
		InvitationsAPI: func(server, token string) (ports.InvitationsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		AccessGrantsAPI: func(server, token string) (ports.AccessGrantsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		ActivityAPI: func(server, token string) (ports.ActivityAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		AuditAPI: func(server, token string) (ports.AuditAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		OperationsAPI: func(server, token string) (ports.OperationsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		SecretInput: secretInput,
		NotificationPreferencesAPI: func(server, token string) (ports.NotificationPreferencesAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		NotificationDevicesAPI: func(server, token string) (ports.NotificationDevicesAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		NotificationsAPI: func(server, token string) (ports.NotificationsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		AttachmentUploads: func(server, token string) (ports.AttachmentUploads, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		AttachmentsAPI: func(server, token string) (ports.AttachmentsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		TagsAPI: func(server, token string) (ports.TagsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		DirectoryLifecycle: func(server, token string) (ports.DirectoryLifecycle, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		TextInput:  textInput,
		InputFiles: inputfiles.Files{Stdin: os.Stdin, StdinTerminal: term.IsTerminal(int(os.Stdin.Fd()))},
		DirectoryWriter: func(server, token string) (ports.DirectoryWriter, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		DirectoryAPI: func(server, token string) (ports.Directory, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		Picker: picker,
		ScopeAPI: func(server, token string) (ports.ScopeCatalog, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		Contexts:   contextStore,
		LabelFiles: labelfiles.Files{},
		LabelsAPI: func(server, token string) (ports.LabelsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		PrintingAPI: func(server, token string) (ports.HumanPrintingAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		API: func(server, token string) (ports.API, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		Auth:        oidcauth.Adapter{HTTP: client, Clock: clock, Output: output, Browser: oidcauth.SystemBrowser{}, AllowLoopbackHTTP: options.AllowLoopbackHTTP},
		Credentials: store, Clock: clock, Output: output, Observer: presentation.SilentObserver{},
	}
	return exit(output, runner.Run(ctx, options))
}
func exit(output ports.Output, err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) {
		output.Error("canceled", "The command was canceled. No further actions will run.")
		return 130
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
  stuffstash context list
  stuffstash context current
  stuffstash context use NAME
  stuffstash context delete NAME
  stuffstash tenants create --name NAME
  stuffstash tenants update --name NAME
  stuffstash inventories create --name NAME
  stuffstash inventories update --input FILE|-
  stuffstash tenants archive|restore|delete [--yes]
  stuffstash inventories archive|restore|delete [--yes]
  stuffstash tenants audit [--limit N --cursor CURSOR]
  stuffstash inventories audit [--limit N --cursor CURSOR]
  stuffstash assets audit ASSET_ID [--limit N]
  stuffstash invitations list [--status STATUS --limit N --cursor CURSOR]
  stuffstash invitations show|cancel|delete ID [--yes]
  stuffstash access-grants create [--input FILE|-] [--yes]
  stuffstash access-grants list [--limit N --cursor CURSOR]
  stuffstash access-grants show|remove PRINCIPAL_ID viewer|editor [--yes]
  stuffstash assets activity ASSET_ID [--view changes|all --limit N --cursor CURSOR]
  stuffstash operations undo|redo OPERATION_ID [--yes]
  stuffstash notification-preferences update [--input FILE|-]
  stuffstash notification-preferences override TYPE_ID [--input FILE|-]
  stuffstash notification-preferences remove-override TYPE_ID --revision N [--yes]
  stuffstash notification-preferences show
  stuffstash notification-preferences initialize [--timezone ZONE | --input FILE|-]
  stuffstash notification-devices register [--input FILE|-]
  stuffstash notification-devices show INSTALLATION_ID
  stuffstash notification-devices remove DEVICE_ID --revision N [--yes]
  stuffstash notifications list [--unread-only --limit N --cursor CURSOR]
  stuffstash notifications show|read|unread ID
  stuffstash notifications unread-count|read-all [--cursor CURSOR]
  stuffstash attachments list ASSET_ID [--limit N --cursor CURSOR]
  stuffstash attachments show ASSET_ID ATTACHMENT_ID
  stuffstash attachments complete-upload ASSET_ID UPLOAD_ID
  stuffstash attachments archive|restore|delete ASSET_ID ATTACHMENT_ID [--yes]
  stuffstash tags list [--limit N --cursor CURSOR]
  stuffstash tags create --name NAME [--key KEY --tag-color HEX]
  stuffstash tags update ID [--name NAME --tag-color HEX | --input FILE]
  stuffstash tags delete ID [--yes]
  stuffstash account show
  stuffstash tenants show [--tenant ID]
  stuffstash inventories show [--tenant ID --inventory ID]
  stuffstash tenants list [--limit N --cursor CURSOR]
  stuffstash inventories list --tenant ID
  stuffstash assets list [--tenant ID --inventory ID --limit N --cursor CURSOR]
                       [--lifecycle active|archived|all --sort id_asc|updated_desc]
  stuffstash assets show ID
  stuffstash assets create --kind item|container|location --title TITLE [--print-label]
  stuffstash assets update ID --title TITLE
  stuffstash assets create --input FILE|-
  stuffstash assets update ID --input FILE|-
  stuffstash assets move ID --parent ID|root
  stuffstash assets archive ID [--yes]
  stuffstash assets restore ID
  stuffstash assets delete ID [--yes]
  stuffstash assets expiration [--mode all|soon|expired --kind KIND --checkout-state STATE]
    [--query TEXT --type-id ID --tag-id ID --location-id ID --from-date YYYY-MM-DD --through-date YYYY-MM-DD]
    [--limit N --cursor CURSOR]
  stuffstash assets checked-out [--limit N --cursor CURSOR]
  stuffstash assets checkout ID [--details TEXT | --input FILE|-]
  stuffstash assets return ID [--details TEXT | --input FILE|-]
  stuffstash assets checkouts ID [--limit N --cursor CURSOR]
  stuffstash assets return-details ASSET_ID CHECKOUT_ID --details TEXT | --input FILE|-
  stuffstash version
  stuffstash labels templates
  stuffstash labels render ASSET_ID --format png|pdf --output PATH [--printer ID | --media-preset ID]
  stuffstash labels resolve LABEL_URL
  stuffstash labels print ASSET_ID [--printer ID --template ID --template-version N]
  stuffstash printers list
  stuffstash printers configure PRINTER_ID --label-size PRESET_ID
  stuffstash printers test PRINTER_ID
  stuffstash print-jobs list [--printer ID]
  stuffstash print-jobs show JOB_ID
  stuffstash print-jobs cancel JOB_ID
  stuffstash print-jobs reprint JOB_ID [--printer ID]
  stuffstash printers discover
  stuffstash printers catalog [--json]
  stuffstash connectors print register --name NAME
  stuffstash connectors print rotate --connector ID
  stuffstash connectors print run --connector ID [--journal-dir PATH]

Context: --server, --tenant, --inventory or STUFF_STASH_CLI_SERVER,
STUFF_STASH_CLI_TENANT, STUFF_STASH_CLI_INVENTORY override the saved context.
Use --context NAME or STUFF_STASH_CLI_CONTEXT to choose a saved context.
Saved resource scope is reused only for the same signed-in account.
Render writes a new private file; existing paths are never overwritten.
Standalone dimensions: --width-mm WIDTH --height-mm HEIGHT (exact catalog geometry).
Finite commands accept --json and --request-id.
Use --idempotency-key only for operations that support it.
Headless credential storage: explicitly set STUFF_STASH_CLI_CREDENTIAL_FILE.
Local printer discovery and catalog export do not need login.
Connector secrets: use the OS store or explicitly set
STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE for headless hosts.
Registration prints a browser approval URL and short code.
The Linux USB worker consumes all assigned printers. Keep its journal directory persistent.
`
