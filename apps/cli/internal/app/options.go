package app

import (
	"flag"
	"io"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Options struct {
	Command                                             []string
	Server                                              string
	Scope                                               ports.Scope
	Page                                                ports.Page
	Title, Kind, Parent, IdempotencyKey, CredentialFile string
	JSON, DeviceCode, AllowLoopbackHTTP, Help           bool
}

func Parse(args []string, getenv func(string) string) (Options, error) {
	o := Options{Server: getenv("STUFF_STASH_CLI_SERVER"), Scope: ports.Scope{Tenant: getenv("STUFF_STASH_CLI_TENANT"), Inventory: getenv("STUFF_STASH_CLI_INVENTORY")}, CredentialFile: getenv("STUFF_STASH_CLI_CREDENTIAL_FILE")}
	if raw := getenv("STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return o, ports.Failure("configuration", "invalid STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP")
		}
		o.AllowLoopbackHTTP = v
	}
	flags := flag.NewFlagSet("stuffstash", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
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
	o.Command = positional
	if o.Page.Limit < 1 {
		return o, ports.Failure("usage", "--limit must be positive")
	}
	return o, nil
}
