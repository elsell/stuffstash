package app

import (
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func customizationFlags(f *flag.FlagSet, o *Options) {
	f.StringVar(&o.DefinitionLevel, "scope", "", "definition scope: household or inventory")
	f.StringVar(&o.FieldType, "field-type", "", "field type: text, number, boolean, date, url, or enum")
	f.Func("description", "asset type description (empty clears it)", func(v string) error { o.TypeDescription = &v; return nil })
	f.BoolFunc("expiration-enabled", "enable asset type expiration (use =false to disable)", func(v string) error {
		b, e := strconv.ParseBool(v)
		if e == nil {
			o.TypeExpiration = &b
		}
		return e
	})
}
func validateCustomizationFlags(f *flag.FlagSet, o Options) error {
	var problem error
	f.Visit(func(v *flag.Flag) {
		if problem != nil {
			return
		}
		if !isCustomization(o) {
			if v.Name == "scope" || v.Name == "field-type" || v.Name == "description" || v.Name == "expiration-enabled" {
				problem = ports.Failure("usage", "Use --"+v.Name+" only with asset-types or field-definitions.")
			}
			return
		}
		allowed := false
		field := false
		switch v.Name {
		case "scope", "server", "tenant", "inventory", "context", "json", "no-input", "color", "request-id", "help":
			allowed = true
		case "limit", "cursor", "lifecycle":
			allowed = customizationList(o)
		case "yes":
			allowed = len(o.Command) > 1 && o.Command[1] != "list" && o.Command[1] != "show"
		case "input":
			allowed = customizationBody(o)
		case "name":
			allowed = customizationBody(o)
			field = true
		case "key":
			allowed = customizationBody(o) && o.Command[1] == "create"
			field = true
		case "field-type":
			allowed = customizationBody(o) && o.Command[0] == "field-definitions" && o.Command[1] == "create"
			field = true
		case "description", "expiration-enabled":
			allowed = customizationBody(o) && o.Command[0] == "asset-types"
			field = true
		}
		if !allowed {
			problem = ports.Failure("usage", "This definition command does not accept --"+v.Name+". See its --help.")
			return
		}
		if field && o.InputPath != "" {
			problem = ports.Failure("usage", "Use either --input or definition field options. Do not combine them.")
		}
		if v.Name == "scope" && o.DefinitionLevel == "" {
			problem = ports.Failure("usage", "Supply --scope household or --scope inventory.")
		}
	})
	return problem
}
