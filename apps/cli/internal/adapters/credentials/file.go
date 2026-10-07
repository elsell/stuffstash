package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// File is an explicit headless alternative; it never silently replaces a keyring.
type File struct{ Path string }

func checkFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ports.Failure("configuration", "The credential path is not a regular file. Select a private regular file for credential storage.")
	}
	if info.Mode().Perm()&0077 != 0 {
		return ports.Failure("configuration", "Other accounts can access the credential file. Set its permissions to 0600.")
	}
	return checkOwner(info)
}
func (f File) Load(_ context.Context, server string) (ports.Session, error) {
	var session ports.Session
	if err := checkFile(f.Path); err != nil {
		if os.IsNotExist(err) {
			return session, ports.ErrNotLoggedIn
		}
		return session, err
	}
	body, err := os.ReadFile(f.Path)
	if err != nil {
		return session, err
	}
	if json.Unmarshal(body, &session) != nil {
		return session, ports.Failure("configuration", "The credential file format is not correct. Run stuffstash login again.")
	}
	if session.Server != server {
		return ports.Session{}, ports.ErrNotLoggedIn
	}
	return session, nil
}
func (f File) Save(_ context.Context, session ports.Session) error {
	if err := checkFile(f.Path); err != nil && !os.IsNotExist(err) {
		return err
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
	body, err := json.Marshal(session)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".stuffstash-session-*")
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
	return os.Rename(file.Name(), f.Path)
}
func (f File) Delete(ctx context.Context, server string) error {
	if _, err := f.Load(ctx, server); err != nil {
		if errors.Is(err, ports.ErrNotLoggedIn) {
			return nil
		}
		return err
	}
	return os.Remove(f.Path)
}
