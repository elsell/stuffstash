package blobstore

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type FileSystemStore struct {
	root string
}

func NewFileSystemStore(root string) FileSystemStore {
	return FileSystemStore{root: root}
}

func (s FileSystemStore) PutBlob(_ context.Context, key media.StorageKey, _ media.ContentType, data []byte) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (s FileSystemStore) GetBlob(_ context.Context, key media.StorageKey) ([]byte, error) {
	path, err := s.pathForKey(key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ports.ErrBlobNotFound
	}
	return data, err
}

func (s FileSystemStore) DeleteBlob(_ context.Context, key media.StorageKey) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (s FileSystemStore) pathForKey(key media.StorageKey) (string, error) {
	root := strings.TrimSpace(s.root)
	if root == "" {
		return "", errors.New("blob storage path is required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	path := filepath.Join(absoluteRoot, filepath.FromSlash(key.String()))
	relative, err := filepath.Rel(absoluteRoot, path)
	if err != nil {
		return "", err
	}
	if relative == "." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || relative == ".." {
		return "", errors.New("blob storage key escapes storage root")
	}
	return path, nil
}
