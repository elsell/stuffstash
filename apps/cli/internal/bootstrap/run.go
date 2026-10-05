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
	if len(args) > 0 && args[0] == "__complete" {
		return exit(presentation.Output{Stdout: stdout, Stderr: stderr}, completionQuery(stdout, args))
	}
	helpOptions, requested, helpErr := app.ParseHelp(args)
	if requested {
		output := presentation.Output{Stdout: stdout, Stderr: stderr}
		if helpErr != nil {
			return exit(output, helpErr)
		}
		return exit(output, writeHelp(stdout, helpOptions.Command))
	}
	if len(helpOptions.Command) > 0 && helpOptions.Command[0] == "completion" {
		options, err := app.Parse(args, func(string) string { return "" })
		output := presentation.Output{Stdout: stdout, Stderr: stderr}
		if err != nil {
			return exit(output, err)
		}
		return exit(output, writeCompletion(stdout, options.Command))
	}
	options, err := app.Parse(args, getenv)
	output := presentation.Output{Stdout: stdout, Stderr: stderr, JSON: options.JSON}

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
	connector := len(options.Command) == 3 && options.Command[0] == "connectors" && options.Command[1] == "print" && (options.Command[2] == "register" || options.Command[2] == "rotate" || options.Command[2] == "run")
	if !connector && !app.IsConsumerInspection(options) {
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
	if app.IsConsumerInspection(options) {
		inspector := app.ConsumerInspector{Credentials: connectorCredentialStore(options), Clock: systemClock{}, Output: output, Observer: presentation.SilentObserver{}, API: func(server, token string) (ports.ConsumerInspectionAPI, error) {
			return httpapi.New(server, token, &http.Client{Timeout: 30 * time.Second}, httpapi.Options{RequestID: options.RequestID})
		}}
		return exit(output, inspector.Run(ctx, options))
	}

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
		EvaluationAPI: func(server, token string) (ports.EvaluationAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		ConnectorInspectionAPI: func(server, token string) (ports.ConnectorInspectionAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		PrintSettingsAPI: func(server, token string) (ports.PrintSettingsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		VoiceProviderAPI: func(server, token string) (ports.VoiceProviderAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		WorkflowsAPI: func(server, token string) (ports.WorkflowsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		ProviderProfilesAPI: func(server, token string) (ports.ProviderProfilesAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		ImportJobsAPI: func(server, token string) (ports.ImportJobsAPI, error) {
			return httpapi.New(server, token, client, httpapi.Options{RequestID: options.RequestID})
		},
		ServerAPI: func(server string) (ports.ServerAPI, error) {
			return httpapi.New(server, "", client, httpapi.Options{RequestID: options.RequestID})
		},
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
