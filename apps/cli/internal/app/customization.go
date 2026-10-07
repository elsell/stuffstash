package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isCustomization(o Options) bool {
	return len(o.Command) > 0 && (o.Command[0] == "asset-types" || o.Command[0] == "field-definitions")
}
func customizationBody(o Options) bool {
	return isCustomization(o) && len(o.Command) > 1 && (o.Command[1] == "create" || o.Command[1] == "update")
}
func customizationList(o Options) bool {
	return isCustomization(o) && len(o.Command) == 2 && o.Command[1] == "list"
}
func validateCustomization(o Options, scoped bool) error {
	c := o.Command
	valid := len(c) == 2 && (c[1] == "list" || c[1] == "create") || len(c) == 3 && c[2] != "" && (c[1] == "show" || c[1] == "update" || c[1] == "archive" || c[1] == "restore" || c[1] == "delete")
	if !valid {
		return ports.Failure("usage", "Use list, show ID, create, update ID, archive ID, restore ID, or delete ID. See this command's --help.")
	}
	if o.DefinitionLevel != "" && o.DefinitionLevel != "household" && o.DefinitionLevel != "inventory" {
		return ports.Failure("usage", "Use --scope household or --scope inventory.")
	}
	if scoped && (o.Scope.Tenant == "" || o.DefinitionLevel == "inventory" && o.Scope.Inventory == "") {
		return ports.Failure("usage", "Select the household and, for inventory scope, an inventory. Use --tenant and --inventory or a saved context.")
	}
	return nil
}
func (r Runner) chooseDefinitionLevel(ctx context.Context, o Options) (Options, error) {
	if !isCustomization(o) || o.DefinitionLevel != "" {
		return o, nil
	}
	if r.Picker == nil || o.JSON || o.NoInput {
		return o, ports.Failure("usage", "Supply --scope household or --scope inventory. Scripts must select the definition scope explicitly.")
	}
	level, err := r.Picker.Pick(ctx, "Definition scope", []ports.Choice{{ID: "household", Label: "Household", Detail: "Shared definitions for this household"}, {ID: "inventory", Label: "Inventory", Detail: "Definitions for one inventory"}})
	if err != nil {
		return o, err
	}
	if level != "household" && level != "inventory" {
		return o, context.Canceled
	}
	o.DefinitionLevel = level
	return o, nil
}
func (r Runner) customizationCommand(ctx context.Context, o Options, token string) error {
	scope := ports.DefinitionScope{Level: ports.DefinitionLevel(o.DefinitionLevel), Scope: o.Scope}
	action := o.Command[1]
	id := ""
	if len(o.Command) == 3 {
		id = o.Command[2]
	}
	if action != "list" && action != "show" {
		target := id
		if target == "" {
			target = "new definition"
		}
		detail := "Server: " + strconv.Quote(o.Server) + ". Scope: " + o.DefinitionLevel + ". Household: " + strconv.Quote(o.Scope.Tenant)
		if o.DefinitionLevel == "inventory" {
			detail += ". Inventory: " + strconv.Quote(o.Scope.Inventory)
		}
		detail += ". Target: " + strconv.Quote(target) + ". Action: " + action + "."
		if err := r.Output.Notice(detail); err != nil {
			return err
		}
		if err := r.confirmAction(ctx, o, "Change "+o.Command[0], action+" definition", detail); err != nil {
			return err
		}
	}
	var result any
	var err error
	if o.Command[0] == "asset-types" {
		if r.AssetTypesAPI == nil {
			return ports.Failure("configuration", "Asset type commands are not available. Update the CLI.")
		}
		api, e := r.AssetTypesAPI(o.Server, token)
		if e != nil {
			return e
		}
		switch action {
		case "list":
			result, err = listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.AssetType], error) {
				return api.AssetTypes(ctx, scope, page, o.Lifecycle)
			})
		case "show":
			result, err = api.AssetType(ctx, scope, id)
		case "delete":
			err = api.DeleteAssetType(ctx, scope, id)
		default:
			result, err = api.ChangeAssetType(ctx, scope, id, ports.TypeAction(action), o.RequestBody)
		}
	} else {
		if r.FieldDefinitionsAPI == nil {
			return ports.Failure("configuration", "Field definition commands are not available. Update the CLI.")
		}
		api, e := r.FieldDefinitionsAPI(o.Server, token)
		if e != nil {
			return e
		}
		switch action {
		case "list":
			result, err = listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.FieldDefinition], error) {
				return api.FieldDefinitions(ctx, scope, page, o.Lifecycle)
			})
		case "show":
			result, err = api.FieldDefinition(ctx, scope, id)
		case "delete":
			err = api.DeleteFieldDefinition(ctx, scope, id)
		default:
			result, err = api.ChangeFieldDefinition(ctx, scope, id, ports.FieldAction(action), o.RequestBody)
		}
	}
	if err != nil {
		if action != "list" && action != "show" {
			var failure *ports.Error
			if errors.As(err, &failure) {
				switch failure.Category {
				case "conflict":
					return ports.Failure("conflict", "The server rejected the definition change. Examine the definition and its change restrictions before you try again.")
				case "network", "protocol", "unavailable", "api":
					return ports.Failure(failure.Category, "The definition change result is unknown. Use this command's list operation to examine definitions before you try again.")
				}
			}
		}
		return err
	}
	if action == "delete" {
		result = map[string]string{"status": "deleted", "id": id}
	}
	r.Observer.Event(ctx, "cli.customization."+action+".completed")
	return r.Output.Result(result)
}
