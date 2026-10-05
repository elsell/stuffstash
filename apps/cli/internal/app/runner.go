package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Runner struct {
	InputFiles      ports.InputFiles
	DirectoryWriter func(string, string) (ports.DirectoryWriter, error)
	DirectoryAPI    func(string, string) (ports.Directory, error)
	Picker          ports.Selector
	ScopeAPI        func(string, string) (ports.ScopeCatalog, error)
	Contexts        contexts.Store
	LabelsAPI       func(string, string) (ports.LabelsAPI, error)
	LabelFiles      ports.LabelFiles
	PrintingAPI     func(string, string) (ports.HumanPrintingAPI, error)
	API             func(string, string) (ports.API, error)
	Auth            ports.Auth
	Credentials     ports.Credentials
	Output          ports.Output
	Clock           ports.Clock
	Observer        ports.Observer
}

func (r Runner) Run(ctx context.Context, o Options) error {
	if o.InputPath != "" && !isDirectoryWrite(o) {
		return ports.Failure("usage", "This command does not accept --input. Remove the option.")
	}
	if len(o.Command) == 0 {
		return ports.Failure("usage", "a command is required; use --help")
	}
	if o.Command[0] == "context" {
		return r.contextCommand(ctx, o.Command)
	}
	switch o.Command[0] {
	case "login":
		if len(o.Command) != 1 {
			return ports.Failure("usage", "login takes no positional arguments")
		}
		api, err := r.API(o.Server, "")
		if err != nil {
			return err
		}
		metadata, err := api.AuthConfig(ctx)
		if err != nil {
			return err
		}
		session, err := r.Auth.Login(ctx, o.Server, metadata, o.DeviceCode)
		if err != nil {
			return err
		}
		if err := r.Credentials.Save(ctx, session); err != nil {
			return err
		}
		r.Observer.Event(ctx, "cli.login.completed")
		return r.Output.Result(map[string]string{"status": "signed in", "server": o.Server})
	case "logout":
		if len(o.Command) != 1 {
			return ports.Failure("usage", "logout takes no positional arguments")
		}
		if err := r.Credentials.Delete(ctx, o.Server); err != nil {
			return err
		}
		if r.Contexts != nil {
			if err := (contexts.Manager{Store: r.Contexts}).ClearServer(ctx, o.Server); err != nil {
				return err
			}
		}
		r.Observer.Event(ctx, "cli.logout.completed")
		return r.Output.Result(map[string]string{"status": "signed out"})
	}
	if err := validateCommandShape(o); err != nil {
		return err
	}
	o, err := r.prepareInput(ctx, o)
	if err != nil {
		return err
	}
	session, err := r.Credentials.Load(ctx, o.Server)
	if err != nil {
		return err
	}
	if !session.ExpiresAt.After(r.Clock.Now().Add(30*time.Second)) || r.Contexts != nil && session.Subject == "" && missingResourceScope(o) {
		session, err = r.Auth.Refresh(ctx, session)
		if err != nil {
			return err
		}
		if err = r.Credentials.Save(ctx, session); err != nil {
			return err
		}
	}
	if isTenantCreate(o) {
		return r.writeDirectory(ctx, o, session.IDToken)
	}
	if isAccountCommand(o) {
		return r.directoryCommand(ctx, o, session.IDToken)
	}
	if r.Contexts != nil {
		config, configErr := r.Contexts.Load(ctx)
		if configErr != nil {
			return configErr
		}
		request := o.Selection
		request.Server = o.Server
		request.Tenant = o.Scope.Tenant
		request.Inventory = o.Scope.Inventory
		resolved, resolveErr := contexts.Resolve(config, request, contexts.Principal(session))
		if resolveErr != nil {
			return resolveErr
		}
		o.Scope = resolved.Scope
	}
	o, err = r.chooseMissingScope(ctx, o, session)
	if err != nil {
		return err
	}
	if err := validateCommand(o); err != nil {
		return err
	}

	if isDirectoryWrite(o) {
		return r.writeDirectory(ctx, o, session.IDToken)
	}
	if isDirectoryCommand(o) {
		return r.directoryCommand(ctx, o, session.IDToken)
	}
	api, err := r.API(o.Server, session.IDToken)
	if err != nil {
		return err
	}
	var result any
	if isLabelCommand(o) {
		if r.LabelsAPI == nil {
			return ports.Failure("configuration", "label API is unavailable")
		}
		labelAPI, labelErr := r.LabelsAPI(o.Server, session.IDToken)
		if labelErr != nil {
			return labelErr
		}
		result, err = executeLabels(ctx, labelAPI, r.LabelFiles, o)
	} else if isPrintingCommand(o) {
		if r.PrintingAPI == nil {
			return ports.Failure("configuration", "printing API is unavailable")
		}
		printingAPI, printErr := r.PrintingAPI(o.Server, session.IDToken)
		if printErr != nil {
			return printErr
		}
		if o.IdempotencyKey == "" && (o.PrintLabel || o.Command[1] == "print" || o.Command[1] == "test" || o.Command[1] == "reprint") {
			var token [16]byte
			if _, err = rand.Read(token[:]); err != nil {
				return err
			}
			o.IdempotencyKey = hex.EncodeToString(token[:])
		}
		if o.IdempotencyKey != "" {
			if err = r.Output.Notice("Print request key: " + o.IdempotencyKey + "; reuse this key and selection if the response is lost."); err != nil {
				return err
			}
		}
		if o.PrintLabel {
			selection, selectErr := selectPrinter(ctx, printingAPI, o)
			if selectErr != nil {
				return selectErr
			}
			result, err = api.CreateAsset(ctx, o.Scope, ports.AssetInput{Kind: o.Kind, Title: o.Title, Parent: o.Parent, PrintLabel: &selection}, o.IdempotencyKey)
		} else {
			result, err = executePrinting(ctx, printingAPI, o)
		}
	} else {
		result, err = execute(ctx, api, o)
	}
	if err != nil {
		return err
	}
	if isLabelCommand(o) {
		r.Observer.Event(ctx, "cli.label.command.completed")
	} else if isPrintingCommand(o) {
		r.Observer.Event(ctx, "cli.print.command.completed")
	} else {
		r.Observer.Event(ctx, "cli.inventory.command.completed")
	}
	return r.Output.Result(result)
}
func validateCommand(o Options) error      { return validateCommandOptions(o, true) }
func validateCommandShape(o Options) error { return validateCommandOptions(o, false) }
func validateCommandOptions(o Options, requireScope bool) error {
	if o.PrintLabel && (len(o.Command) != 2 || o.Command[0] != "assets" || o.Command[1] != "create") {
		return ports.Failure("usage", "--print-label is only available for assets create")
	}
	if isLabelCommand(o) {
		return validateLabelCommandOptions(o, requireScope)
	}
	if !o.PrintLabel && isPrintingCommand(o) {
		return validatePrintingCommandOptions(o, requireScope)
	}
	if isDirectoryCommand(o) || isDirectoryWrite(o) {
		if requireScope && missingResourceScope(o) {
			return ports.Failure("usage", "Supply the required scope with --tenant and, for inventory commands, --inventory.")
		}
		return nil
	}
	if len(o.Command) < 2 {
		return ports.Failure("usage", "expected inventories list or assets <action>")
	}
	if requireScope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "choose a tenant with --tenant or STUFF_STASH_CLI_TENANT")
	}
	if o.Command[0] == "inventories" && o.Command[1] == "list" && len(o.Command) == 2 {
		return nil
	}
	if o.Command[0] != "assets" {
		return ports.Failure("usage", "unknown command; use --help")
	}
	if requireScope && o.Scope.Inventory == "" {
		return ports.Failure("usage", "choose an inventory with --inventory or STUFF_STASH_CLI_INVENTORY")
	}
	switch o.Command[1] {
	case "list":
		if len(o.Command) == 2 {
			return nil
		}
	case "create":
		if len(o.Command) == 2 && o.Title != "" && o.Kind != "" {
			return nil
		}
	case "show", "archive", "restore":
		if len(o.Command) == 3 {
			return nil
		}
	case "update":
		if len(o.Command) == 3 && o.Title != "" {
			return nil
		}
	case "move":
		if len(o.Command) == 3 && o.Parent != "" {
			return nil
		}
	}
	return ports.Failure("usage", "invalid asset command arguments; use --help")
}
func execute(ctx context.Context, api ports.API, o Options) (any, error) {
	if o.Command[0] == "inventories" {
		return api.Inventories(ctx, o.Scope, o.Page)
	}
	action := o.Command[1]
	if action == "list" {
		return api.Assets(ctx, o.Scope, o.Page)
	}
	id := ""
	if len(o.Command) > 2 {
		id = o.Command[2]
	}
	if action == "show" {
		return api.Asset(ctx, o.Scope, id)
	}
	key := o.IdempotencyKey
	if key == "" {
		var bytes [16]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			return nil, err
		}
		key = hex.EncodeToString(bytes[:])
	}
	switch action {
	case "create":
		return api.CreateAsset(ctx, o.Scope, ports.AssetInput{Kind: o.Kind, Title: o.Title, Parent: o.Parent}, key)
	case "update":
		return api.UpdateAsset(ctx, o.Scope, id, ports.AssetChange{Title: &o.Title}, key)
	case "move":
		change := ports.AssetChange{Parent: &o.Parent}
		if o.Parent == "root" {
			change = ports.AssetChange{MoveToRoot: true}
		}
		return api.UpdateAsset(ctx, o.Scope, id, change, key)
	case "archive", "restore":
		return api.SetArchived(ctx, o.Scope, id, action == "archive", key)
	}
	return nil, ports.Failure("usage", "unknown action")
}

func isTenantList(o Options) bool {
	return len(o.Command) == 2 && o.Command[0] == "tenants" && o.Command[1] == "list"
}
