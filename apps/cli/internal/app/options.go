package app

import (
	"flag"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Options struct {
	AllInventories bool

	FilePath, Transfer, Variant string

	DefinitionLevel, FieldType string
	TypeDescription            *string
	TypeExpiration             *bool

	PrinterAdapterID, PrinterPresetVersion                                string
	ExpectedMediaFingerprint, PreviewFingerprint                          string
	InvitationStatus                                                      string
	ActivityView                                                          ports.ActivityView
	Timezone                                                              string
	Revision                                                              int64
	UnreadOnly                                                            bool
	Expiration                                                            ports.ExpirationQuery
	Details                                                               *string
	Lifecycle, Sort                                                       string
	TagColor, TagKey                                                      *string
	Yes                                                                   bool
	RequestID                                                             string
	InputPath                                                             string
	RequestBody                                                           []byte
	Color                                                                 string
	Selection                                                             contexts.Selection
	NoInput                                                               bool
	Format, OutputPath, MediaPreset                                       string
	WidthMM, HeightMM                                                     float64
	PrinterID, TemplateID, LabelSize                                      string
	TemplateVersion                                                       uint
	Copies                                                                int
	PrintLabel, ShowReference, ShowReferenceSet                           bool
	Command                                                               []string
	ConnectorID, ConnectorName, ConnectorCredentialFile, JournalDirectory string
	Server                                                                string
	Scope                                                                 ports.Scope
	Page                                                                  ports.Page
	Title, Kind, Parent, IdempotencyKey, CredentialFile                   string
	JSON, DeviceCode, AllowLoopbackHTTP, Help                             bool
}

func Parse(args []string, getenv func(string) string) (Options, error) {
	o := Options{ConnectorID: getenv("STUFF_STASH_CLI_CONNECTOR_ID"), ConnectorCredentialFile: getenv("STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE"), JournalDirectory: getenv("STUFF_STASH_CLI_PRINT_STATE_DIRECTORY"), Server: getenv("STUFF_STASH_CLI_SERVER"), Scope: ports.Scope{Tenant: getenv("STUFF_STASH_CLI_TENANT"), Inventory: getenv("STUFF_STASH_CLI_INVENTORY")}, CredentialFile: getenv("STUFF_STASH_CLI_CREDENTIAL_FILE")}
	if raw := getenv("STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return o, ports.Failure("configuration", "invalid STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP")
		}
		o.AllowLoopbackHTTP = v
	}
	environment := contexts.Selection{Context: getenv("STUFF_STASH_CLI_CONTEXT"), Server: o.Server, Tenant: o.Scope.Tenant, Inventory: o.Scope.Inventory}
	o.Selection.Context = environment.Context
	flags := optionFlags(&o)
	flagArgs, positional, _, splitErr := partitionOptions(args, flags)
	if splitErr != nil {
		return o, splitErr
	}
	if err := parseOptionValues(flags, flagArgs); err != nil {
		return o, err
	}
	explicit := contexts.Selection{}
	var emptyScopeFlag string
	flags.Visit(func(f *flag.Flag) {
		if (f.Name == "server" || f.Name == "tenant" || f.Name == "inventory" || f.Name == "context") && f.Value.String() == "" {
			emptyScopeFlag = f.Name
		}
		switch f.Name {
		case "context":
			explicit.Context = o.Selection.Context
		case "server":
			explicit.Server = o.Server
		case "tenant":
			explicit.Tenant = o.Scope.Tenant
		case "inventory":
			explicit.Inventory = o.Scope.Inventory
		}
		if f.Name == "show-reference" {
			o.ShowReferenceSet = true
		}
	})
	if emptyScopeFlag != "" {
		return o, ports.Failure("usage", "The --"+emptyScopeFlag+" value is empty. Supply a value or remove the option.")
	}
	o.Selection = contexts.Overlay(environment, explicit)
	o.Server = o.Selection.Server
	o.Scope = ports.Scope{Tenant: o.Selection.Tenant, Inventory: o.Selection.Inventory}
	if o.Copies < 1 {
		return o, ports.Failure("usage", "Supply --copies with an integer of at least 1.")
	}
	if o.TemplateVersion > uint(^uint32(0)) {
		return o, ports.Failure("usage", "Supply --template-version with an integer from 0 through 4294967295; 0 uses the default.")
	}
	if o.Color != "auto" && o.Color != "always" && o.Color != "never" {
		return o, ports.Failure("usage", "Use --color auto, --color always, or --color never.")
	}
	if strings.IndexFunc(o.RequestID, func(r rune) bool { return r < 32 || r > 126 }) >= 0 {
		return o, ports.Failure("usage", "The request ID contains invalid characters. Use printable ASCII characters.")
	}
	o.Command = positional
	if err := validateCustomizationFlags(flags, o); err != nil {
		return o, err
	}
	if isSearch(o) {
		return o, validateSearchFlags(o, flags)
	}
	if o.AllInventories {
		return o, ports.Failure("usage", "Use --all-inventories only with assets search.")
	}

	if err := binaryFlags(flags, o); err != nil {
		return o, err
	}
	if err := labelRenderFlags(o, flags); err != nil {
		return o, err
	}
	if err := pairingApprovalFlags(flags, o); err != nil {
		return o, err
	}
	if err := validatePrinterAdministrationFlags(o, flags); err != nil {
		return o, err
	}
	if err := validateConsumerInspection(flags, o); err != nil {
		return o, err
	}
	if err := validatePrintSubmissionFlags(flags, o); err != nil {
		return o, err
	}
	if isEvaluationCommand(o) {
		unsupported := ""
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "server", "tenant", "inventory", "context", "credential-file", "allow-loopback-http", "json", "no-input", "request-id", "color", "help":
			case "input", "yes":
				if !isEvaluationCancellation(o) && !isEvaluationWrite(o) {
					unsupported = f.Name
				}
			case "limit", "cursor":
				if len(o.Command) < 3 || o.Command[2] != "list" {
					unsupported = f.Name
				}
			default:
				unsupported = f.Name
			}
		})
		if unsupported != "" {
			return o, ports.Failure("usage", "Evaluation commands do not accept --"+unsupported+" for this command.")
		}
	}
	if len(o.Command) > 1 && o.Command[0] == "labels" && (o.Command[1] == "show" || o.Command[1] == "assign") {
		invalid := ""
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "format", "output", "media-preset", "width-mm", "height-mm", "printer", "template", "template-version", "copies", "show-reference", "label-size":
				invalid = f.Name
			}
		})
		if invalid != "" {
			return o, ports.Failure("usage", "Label identity commands do not accept --"+invalid+". Use labels render or labels print for that option.")
		}
	}
	if isTelemetry(o) {
		unsupported := ""
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "context", "server", "json", "no-input", "request-id", "color", "help", "allow-loopback-http", "credential-file", "input", "yes":
			default:
				unsupported = f.Name
			}
		})
		if unsupported != "" {
			return o, ports.Failure("usage", "Telemetry submission does not accept --"+unsupported+". Remove the option.")
		}
	}
	if isVoiceProviderCommand(o) {
		unsupported := ""
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "server", "tenant", "inventory", "context", "json", "no-input", "request-id", "color", "help", "allow-loopback-http", "credential-file":
			case "input", "yes":
				if !isVoiceProviderUpdate(o) {
					unsupported = f.Name
				}
			default:
				unsupported = f.Name
			}
		})
		if unsupported != "" {
			return o, ports.Failure("usage", "This voice provider command does not accept --"+unsupported+". Remove the option.")
		}
	}
	if isWorkflowCommand(o) {
		unsupported := ""
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "server", "tenant", "inventory", "context", "json", "no-input", "request-id", "color", "help", "allow-loopback-http", "credential-file":
			case "input", "yes":
				if !isWorkflowWrite(o) {
					unsupported = f.Name
				}
			case "limit", "cursor":
				if !workflowList(o) {
					unsupported = f.Name
				}
			default:
				unsupported = f.Name
			}
		})
		if unsupported != "" {
			return o, ports.Failure("usage", "This workflow command does not accept --"+unsupported+". Remove the option.")
		}
	}
	if isProviderProfileCommand(o) {
		unsupported := ""
		flags.Visit(func(f *flag.Flag) {
			if f.Name == "limit" {
				unsupported = f.Name
			}
		})
		if unsupported != "" {
			return o, ports.Failure("usage", "Provider profile reads do not support --limit. Remove the option.")
		}
	}
	if o.Timezone != "" && !(isPreferenceCommand(o) && len(o.Command) > 1 && o.Command[1] == "initialize") {
		return o, ports.Failure("usage", "Use --timezone with notification-preferences initialize.")
	}
	if o.Revision != -1 && !isNotificationDeviceCommand(o) && !isPreferenceCommand(o) {
		return o, ports.Failure("usage", "Use --revision with notification-devices remove.")
	}
	if o.InvitationStatus != "" && !isInvitationCommand(o) && !IsConsumerInspection(o) {
		return o, ports.Failure("usage", "Use --status with invitations list.")
	}
	if o.ActivityView != "" && !isActivityCommand(o) {
		return o, ports.Failure("usage", "Use --view with assets activity.")
	}
	if o.UnreadOnly && !isNotificationCommand(o) {
		return o, ports.Failure("usage", "Use --unread-only with notifications list.")
	}
	if err := validateExpiration(o); err != nil {
		return o, err
	}
	if o.Details != nil && !isCheckoutWrite(o) {
		return o, ports.Failure("usage", "Use --details only with checkout, return, or return-details.")
	}
	if o.Lifecycle != "" || o.Sort != "" {
		if !(len(o.Command) == 2 && o.Command[0] == "assets" && o.Command[1] == "list") && !(customizationList(o) && o.Sort == "") {
			return o, ports.Failure("usage", "Use --lifecycle and --sort only with assets list.")
		}
		if o.Lifecycle != "" && o.Lifecycle != "active" && o.Lifecycle != "archived" && o.Lifecycle != "all" {
			return o, ports.Failure("usage", "Use --lifecycle active, archived, or all.")
		}
		if o.Sort != "" && o.Sort != "id_asc" && o.Sort != "updated_desc" {
			return o, ports.Failure("usage", "Use --sort id_asc or updated_desc.")
		}
	}

	if (o.TagColor != nil || o.TagKey != nil && !isCustomization(o)) && !isTagCommand(o) {
		return o, ports.Failure("usage", "Use --tag-color and --key only with tag commands.")
	}
	if o.InputPath != "" && !acceptsBody(o) {
		return o, ports.Failure("usage", "This command does not accept --input. Remove the option.")
	}
	if o.Page.Limit < 0 || o.Page.Limit == 0 && !(len(o.Command) == 2 && o.Command[0] == "tags" && o.Command[1] == "list") {
		return o, ports.Failure("usage", "--limit must be positive")
	}
	return o, nil
}
