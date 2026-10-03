package app

import (
	"flag"
	"io"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Options struct {
	PrinterID, TemplateID                                                 string
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
	flags := flag.NewFlagSet("stuffstash", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&o.PrinterID, "printer", "", "registered printer destination")
	flags.StringVar(&o.TemplateID, "template", "", "label template ID")
	flags.UintVar(&o.TemplateVersion, "template-version", 0, "label template version")
	flags.IntVar(&o.Copies, "copies", 1, "number of label copies")
	flags.BoolVar(&o.PrintLabel, "print-label", false, "create a label job with the new asset")
	flags.BoolVar(&o.ShowReference, "show-reference", false, "show label reference")
	flags.StringVar(&o.ConnectorID, "connector", o.ConnectorID, "registered connector ID")
	flags.StringVar(&o.ConnectorName, "name", "", "connector name")
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
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "show-reference" {
			o.ShowReferenceSet = true
		}
	})
	if o.Copies < 1 || o.TemplateVersion > uint(^uint32(0)) {
		return o, ports.Failure("usage", "invalid copies or template version")
	}
	o.Command = positional
	if o.Page.Limit < 1 {
		return o, ports.Failure("usage", "--limit must be positive")
	}
	return o, nil
}
