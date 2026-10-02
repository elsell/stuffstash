package gormstore

import (
	"context"
	"database/sql"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

var _ ports.ArchiveSnapshotRepository = Store{}

func (s Store) WithArchiveSnapshot(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, read func(ports.ArchiveSnapshotSources) error) error {
	if tenantID.String() == "" || inventoryID.String() == "" || read == nil {
		return ports.ErrArchiveJobScope
	}
	options := &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead}
	if s.db.Dialector.Name() == "sqlite" {
		options.Isolation = sql.LevelSerializable
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scoped := NewStore(tx)
		return read(ports.ArchiveSnapshotSources{Inventories: scoped, Assets: scoped, Tags: scoped, Checkouts: scoped, Attachments: scoped, Types: scoped, Fields: scoped})
	}, options)
}
