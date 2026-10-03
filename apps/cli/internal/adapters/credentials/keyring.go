package credentials

import (
	"context"
	"encoding/json"
	"errors"

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
		return ports.Session{}, errors.New("OS credential store unavailable; headless hosts can explicitly configure STUFF_STASH_CLI_CREDENTIAL_FILE")
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
		return errors.New("could not save to OS credential store")
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
