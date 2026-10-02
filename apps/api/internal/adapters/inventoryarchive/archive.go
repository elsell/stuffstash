// Package inventoryarchive implements the versioned portable ZIP container.
// Inventory semantics and destination authorization are validated by callers.
package inventoryarchive

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"time"
)

var ErrInvalid = errors.New("invalid inventory archive")
var ErrLimit = errors.New("inventory archive exceeds configured limit")

const format = "stuffstash.inventory"
const version = 1
const manifestName = "manifest.json"
const inventoryName = "inventory.json"

type Selection struct {
	Photos     bool `json:"photos"`
	OtherFiles bool `json:"otherFiles"`
}
type Media struct {
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"sizeBytes"`
}
type Limits struct {
	CompressedBytes, ExpandedBytes, MetadataBytes, EntryBytes int64
	Entries                                                   int
}

func (l Limits) valid() bool {
	return l.CompressedBytes > 0 && l.ExpandedBytes > 0 && l.MetadataBytes > 0 && l.EntryBytes > 0 && l.Entries > 0
}

type Source interface {
	Open(context.Context, string) (io.ReadCloser, error)
}

type manifest struct {
	Format     string           `json:"format"`
	Version    int              `json:"archiveVersion"`
	ExportedAt time.Time        `json:"exportedAt"`
	Selection  *selectionRecord `json:"selection"`
	Inventory  Media            `json:"inventory"`
	Media      []Media          `json:"media"`
}
type selectionRecord struct {
	Photos     *bool `json:"photos"`
	OtherFiles *bool `json:"otherFiles"`
}

type Archive struct {
	InventoryJSON []byte
	Selection     Selection
	ExportedAt    time.Time
	Media         []Media
	files         map[string]*zip.File
}

// Open reads a previously validated content entry. The ReaderAt supplied to Read
// must remain open and immutable throughout the archive's lifetime.
func (a *Archive) Open(hash string) (io.ReadCloser, error) {
	f, ok := a.files[hash]
	if !ok {
		return nil, ErrInvalid
	}
	return f.Open()
}
