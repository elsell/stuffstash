package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"github.com/zalando/go-keyring"
)

type Keyring struct{}

const service = "Stuff Stash CLI"

func (Keyring) Load(_ context.Context, server string) (ports.Session, error) {
	value, err := keyring.Get(service, server)
	if errors.Is(err, keyring.ErrNotFound) {
		return ports.Session{}, ports.ErrNotLoggedIn
	}
	if err != nil {
		return ports.Session{}, credentialStoreError("read")
	}
	var s ports.Session
	if json.Unmarshal([]byte(value), &s) != nil || s.Server != server {
		return ports.Session{}, ports.ErrNotLoggedIn
	}
	return s, nil
}
func (Keyring) Save(_ context.Context, s ports.Session) error {
	value, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if keyring.Set(service, s.Server, string(value)) != nil {
		return credentialStoreError("save")
	}
	return nil
}
func (Keyring) Delete(_ context.Context, server string) error {
	err := keyring.Delete(service, server)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func credentialStoreError(action string) error {
	message := "Cannot " + action + " your session in the system credential store. Unlock the store and run stuffstash login again."
	if runtime.GOOS != "windows" {
		message += " You can also set STUFF_STASH_CLI_CREDENTIAL_FILE to a private file path before login."
	}
	return errors.New(message)
}
