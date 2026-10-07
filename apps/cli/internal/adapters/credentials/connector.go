package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"github.com/zalando/go-keyring"
)

const connectorService = "Stuff Stash Print Connector"

type ConnectorFile struct{ Path string }
type ConnectorKeyring struct{}

func validRegistration(r ports.ConnectorRegistration) bool {
	return r.Server != "" && r.TenantID != "" && r.InventoryID != "" && r.ConnectorID != "" && r.Credential != "" && !r.ExpiresAt.IsZero()
}
func decodeRegistration(body []byte) (ports.ConnectorRegistration, error) {
	var r ports.ConnectorRegistration
	if json.Unmarshal(body, &r) != nil || !validRegistration(r) {
		return r, ports.Failure("configuration", "The stored connector registration is not correct. Pair the connector again.")
	}
	return r, nil
}
func (f ConnectorFile) read() (ports.ConnectorRegistration, error) {
	if err := checkFile(f.Path); err != nil {
		if os.IsNotExist(err) {
			return ports.ConnectorRegistration{}, ports.ErrConnectorNotRegistered
		}
		return ports.ConnectorRegistration{}, err
	}
	body, err := os.ReadFile(f.Path)
	if err != nil {
		return ports.ConnectorRegistration{}, err
	}
	registration, err := decodeRegistration(body)
	if err != nil {
		return registration, ports.Failure("configuration", "The stored connector registration is not correct. Set STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE to a new private file path. Pair the connector again.")
	}
	return registration, nil
}
func (f ConnectorFile) Load(_ context.Context, server, id string) (ports.ConnectorRegistration, error) {
	r, err := f.read()
	if err != nil {
		return r, err
	}
	if r.Server != server || r.ConnectorID != id {
		return ports.ConnectorRegistration{}, ports.ErrConnectorNotRegistered
	}
	return r, nil
}
func (f ConnectorFile) Save(_ context.Context, r ports.ConnectorRegistration) error {
	if !validRegistration(r) {
		return ports.Failure("configuration", "The connector registration does not contain all required fields. Examine the connector registration with the administrator.")
	}
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm()&0022 != 0 {
		return ports.Failure("configuration", "The credential storage path must be a directory that other accounts cannot write to. Select a private directory.")
	}
	if err := checkOwner(info); err != nil {
		return err
	}
	unlock, err := lockConnectorFile(f.Path + ".lock")
	if err != nil {
		return err
	}
	defer unlock()
	prior, err := f.read()
	if err != nil && !errors.Is(err, ports.ErrConnectorNotRegistered) {
		return err
	}
	if err == nil && (prior.Server != r.Server || prior.ConnectorID != r.ConnectorID) {
		return ports.Failure("configuration", "The credential file belongs to another connector. Set STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE to a new private file path.")
	}
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".stuffstash-connector-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(body); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(file.Name(), f.Path); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
func (f ConnectorFile) Delete(ctx context.Context, server, id string) error {
	unlock, err := lockConnectorFile(f.Path + ".lock")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer unlock()
	if _, err := f.Load(ctx, server, id); err != nil {
		if errors.Is(err, ports.ErrConnectorNotRegistered) {
			return nil
		}
		return err
	}
	return os.Remove(f.Path)
}
func connectorKey(server, id string) string { return server + "\n" + id }
func (ConnectorKeyring) Load(_ context.Context, server, id string) (ports.ConnectorRegistration, error) {
	value, err := keyring.Get(connectorService, connectorKey(server, id))
	if errors.Is(err, keyring.ErrNotFound) {
		return ports.ConnectorRegistration{}, ports.ErrConnectorNotRegistered
	}
	if err != nil {
		return ports.ConnectorRegistration{}, connectorStoreError("read", runtime.GOOS)
	}
	r, err := decodeRegistration([]byte(value))
	if err != nil {
		return r, err
	}
	if r.Server != server || r.ConnectorID != id {
		return ports.ConnectorRegistration{}, ports.ErrConnectorNotRegistered
	}
	return r, nil
}
func (ConnectorKeyring) Save(_ context.Context, r ports.ConnectorRegistration) error {
	if !validRegistration(r) {
		return ports.Failure("configuration", "The connector registration does not contain all required fields. Examine the connector registration with the administrator.")
	}
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if keyring.Set(connectorService, connectorKey(r.Server, r.ConnectorID), string(body)) != nil {
		return connectorStoreError("save", runtime.GOOS)
	}
	return nil
}
func (ConnectorKeyring) Delete(_ context.Context, server, id string) error {
	err := keyring.Delete(connectorService, connectorKey(server, id))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
