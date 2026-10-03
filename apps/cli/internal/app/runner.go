package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Runner struct {
	API         func(string, string) (ports.API, error)
	Auth        ports.Auth
	Credentials ports.Credentials
	Output      ports.Output
	Clock       ports.Clock
	Observer    ports.Observer
}

func (r Runner) Run(ctx context.Context, o Options) error {
	if len(o.Command) == 0 {
		return ports.Failure("usage", "a command is required; use --help")
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
		r.Observer.Event(ctx, "cli.logout.completed")
		return r.Output.Result(map[string]string{"status": "signed out"})
	}
	if err := validateCommand(o); err != nil {
		return err
	}
	session, err := r.Credentials.Load(ctx, o.Server)
	if err != nil {
		return err
	}
	if !session.ExpiresAt.After(r.Clock.Now().Add(30 * time.Second)) {
		session, err = r.Auth.Refresh(ctx, session)
		if err != nil {
			return err
		}
		if err = r.Credentials.Save(ctx, session); err != nil {
			return err
		}
	}
	api, err := r.API(o.Server, session.IDToken)
	if err != nil {
		return err
	}
	result, err := execute(ctx, api, o)
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.inventory.command.completed")
	return r.Output.Result(result)
}
func validateCommand(o Options) error {
	if len(o.Command) < 2 {
		return ports.Failure("usage", "expected inventories list or assets <action>")
	}
	if o.Scope.Tenant == "" {
		return ports.Failure("usage", "choose a tenant with --tenant or STUFF_STASH_CLI_TENANT")
	}
	if o.Command[0] == "inventories" && o.Command[1] == "list" && len(o.Command) == 2 {
		return nil
	}
	if o.Command[0] != "assets" {
		return ports.Failure("usage", "unknown command; use --help")
	}
	if o.Scope.Inventory == "" {
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
