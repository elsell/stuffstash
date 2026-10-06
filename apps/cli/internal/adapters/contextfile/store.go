package contextfile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"os"
	"path/filepath"
)

const maxConfigBytes = 1024 * 1024

type Store struct{ Path string }

func configError() error {
	return ports.Failure("configuration", "Cannot read the context file. Check its format and access permissions.")
}
func configSaveError() error {
	return ports.Failure("configuration", "Cannot save the context file. Check directory permissions and available disk space.")
}
func (s Store) Load(ctx context.Context) (contexts.Config, error) {
	if err := ctx.Err(); err != nil {
		return contexts.Config{}, err
	}
	root, err := os.OpenRoot(filepath.Dir(s.Path))
	if os.IsNotExist(err) {
		return contexts.Config{Version: contexts.Version}, nil
	}
	if err != nil {
		return contexts.Config{}, configError()
	}
	defer root.Close()
	directory, err := root.Open(".")
	if err != nil {
		return contexts.Config{}, configError()
	}
	safe := secureDirectory(directory)
	directory.Close()
	if !safe {
		return contexts.Config{}, configError()
	}
	return load(root, filepath.Base(s.Path))
}
func load(root *os.Root, name string) (contexts.Config, error) {
	empty := contexts.Config{Version: contexts.Version}
	info, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return empty, nil
	}
	if err != nil || !privateFile(info) {
		return empty, configError()
	}
	file, err := root.Open(name)
	if err != nil {
		return empty, configError()
	}
	defer file.Close()
	actual, err := file.Stat()
	if err != nil || !os.SameFile(info, actual) || !secureFile(file) {
		return empty, configError()
	}
	body, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
	if err != nil || len(body) > maxConfigBytes {
		return empty, configError()
	}
	var config contexts.Config
	if json.Unmarshal(body, &config) != nil || config.Version != contexts.Version {
		return empty, configError()
	}
	if err := contexts.Validate(config); err != nil {
		return empty, err
	}
	return config, nil
}
func (s Store) Update(ctx context.Context, change func(*contexts.Config) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := filepath.Dir(s.Path)
	if err := prepareDirectory(dir); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return configSaveError()
	}
	defer root.Close()
	directory, err := root.Open(".")
	if err != nil {
		return configSaveError()
	}
	safe := secureDirectory(directory)
	directory.Close()
	if !safe {
		return configSaveError()
	}
	unlock, err := lock(ctx, root, filepath.Base(s.Path)+".lock")
	if err != nil {
		return err
	}
	defer unlock()
	config, err := load(root, filepath.Base(s.Path))
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err = change(&config); err != nil {
		return err
	}
	config.Version = contexts.Version
	if err := contexts.Validate(config); err != nil {
		return err
	}
	body, err := json.MarshalIndent(config, "", "  ")
	if err != nil || len(body) > maxConfigBytes {
		return configSaveError()
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	var suffix [12]byte
	if _, err = rand.Read(suffix[:]); err != nil {
		return configSaveError()
	}
	name := ".contexts-" + hex.EncodeToString(suffix[:]) + ".tmp"
	file, err := newPrivateFile(root, name)
	if err != nil {
		return configSaveError()
	}
	defer root.Remove(name)
	if !secureFile(file) {
		file.Close()
		return configSaveError()
	}
	_, err = file.Write(append(body, '\n'))
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		return configSaveError()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err = root.Rename(name, filepath.Base(s.Path)); err != nil {
		return configSaveError()
	}
	return nil
}
