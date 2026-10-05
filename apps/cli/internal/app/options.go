package app

import (
	"flag"
	"io"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Options struct {
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
	flags := flag.NewFlagSet("stuffstash", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&o.InputPath, "input", "", "JSON request file, or - for stdin")
	flags.StringVar(&o.Color, "color", "auto", "color output: auto, always, or never")
	flags.StringVar(&o.Selection.Context, "context", environment.Context, "saved context name")
	flags.BoolVar(&o.NoInput, "no-input", false, "do not ask for input")
	flags.StringVar(&o.Format, "format", "png", "label file format: png or pdf")
	flags.StringVar(&o.OutputPath, "output", "", "new private label file path")
	flags.StringVar(&o.MediaPreset, "media-preset", "", "authorized media preset ID")
	flags.Float64Var(&o.WidthMM, "width-mm", 0, "catalog physical label width in millimeters")
	flags.Float64Var(&o.HeightMM, "height-mm", 0, "catalog physical label height in millimeters")
	flags.StringVar(&o.LabelSize, "label-size", "", "supported printer media preset ID")
	flags.StringVar(&o.PrinterID, "printer", "", "registered printer destination")
	flags.StringVar(&o.TemplateID, "template", "", "label template ID")
	flags.UintVar(&o.TemplateVersion, "template-version", 0, "label template version")
	flags.IntVar(&o.Copies, "copies", 1, "number of label copies")
	flags.BoolVar(&o.PrintLabel, "print-label", false, "create a label job with the new asset")
	flags.BoolVar(&o.ShowReference, "show-reference", false, "show label reference")
	flags.StringVar(&o.ConnectorID, "connector", o.ConnectorID, "registered connector ID")
	flags.StringVar(&o.ConnectorName, "name", "", "household, inventory, or connector name")
	flags.StringVar(&o.JournalDirectory, "journal-dir", o.JournalDirectory, "persistent print recovery directory")
	flags.StringVar(&o.Server, "server", o.Server, "Stuff Stash API URL")
	flags.StringVar(&o.Scope.Tenant, "tenant", o.Scope.Tenant, "tenant ID")
	flags.StringVar(&o.Scope.Inventory, "inventory", o.Scope.Inventory, "inventory ID")
	flags.StringVar(&o.Title, "title", "", "asset title")
	flags.StringVar(&o.Kind, "kind", "", "asset kind")
	flags.StringVar(&o.Parent, "parent", "", "parent ID or root")
	flags.StringVar(&o.IdempotencyKey, "idempotency-key", "", "logical mutation key")
	flags.Int64Var(&o.Page.Limit, "limit", 50, "page size")
	flags.StringVar(&o.Page.Cursor, "cursor", "", "page cursor")
	flags.BoolVar(&o.JSON, "json", false, "JSON output")
	flags.BoolVar(&o.DeviceCode, "device-code", false, "sign in from another device")
	flags.BoolVar(&o.Help, "help", false, "show help")
	// Standard flag parsing stops at the first positional argument. Partition using
	// its registered flag definitions so flags work naturally after command names.
	var flagArgs, positional []string
	for i := 0; i < len(args); i++ {
		word := args[i]
		if word == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(word, "-") {
			positional = append(positional, word)
			continue
		}
		name, _, hasValue := strings.Cut(strings.TrimLeft(word, "-"), "=")
		definition := flags.Lookup(name)
		if definition == nil {
			return o, ports.Failure("usage", "unknown option: "+name)
		}
		flagArgs = append(flagArgs, word)
		boolean, ok := definition.Value.(interface{ IsBoolFlag() bool })
		if !hasValue && (!ok || !boolean.IsBoolFlag()) {
			i++
			if i == len(args) {
				return o, ports.Failure("usage", "missing value for --"+name)
			}
			flagArgs = append(flagArgs, args[i])
		}
	}
	if err := flags.Parse(flagArgs); err != nil {
		return o, ports.Failure("usage", "invalid command option")
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
	if o.Copies < 1 || o.TemplateVersion > uint(^uint32(0)) {
		return o, ports.Failure("usage", "invalid copies or template version")
	}
	if o.Color != "auto" && o.Color != "always" && o.Color != "never" {
		return o, ports.Failure("usage", "Use --color auto, --color always, or --color never.")
	}
	o.Command = positional
 if o.InputPath!="" && !isDirectoryWrite(o){return o,ports.Failure("usage","This command does not accept --input. Remove the option.")}
	if o.Page.Limit < 1 {
		return o, ports.Failure("usage", "--limit must be positive")
	}
	return o, nil
}
