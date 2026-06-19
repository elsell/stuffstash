package dbmigrations

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stuffstash/stuff-stash/migrations"
	"gorm.io/gorm"
)

type Runner struct {
	db *gorm.DB
}

type Status struct {
	Version uint
	Latest  uint
	Dirty   bool
	Empty   bool
}

func NewRunner(db *gorm.DB) Runner {
	return Runner{db: db}
}

func (r Runner) Up() error {
	instance, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer closeMigrate(instance)

	if err := instance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (r Runner) Status() (Status, error) {
	instance, err := r.newMigrate()
	if err != nil {
		return Status{}, err
	}
	defer closeMigrate(instance)

	latest, err := LatestVersion()
	if err != nil {
		return Status{}, err
	}
	version, dirty, err := instance.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return Status{Latest: latest, Empty: true}, nil
	}
	if err != nil {
		return Status{}, err
	}
	return Status{Version: version, Latest: latest, Dirty: dirty}, nil
}

func LatestVersion() (uint, error) {
	names, err := fs.Glob(migrations.Files, "*.up.sql")
	if err != nil {
		return 0, err
	}
	var latest uint64
	for _, name := range names {
		versionText, _, ok := strings.Cut(filepath.Base(name), "_")
		if !ok {
			continue
		}
		version, err := strconv.ParseUint(versionText, 10, 64)
		if err != nil {
			return 0, err
		}
		if version > latest {
			latest = version
		}
	}
	return uint(latest), nil
}

func (r Runner) newMigrate() (*migrate.Migrate, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	databaseDriver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	sourceDriver, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return nil, err
	}
	instance, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseDriver)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func closeMigrate(instance *migrate.Migrate) {
	// The migrate database driver is built from a shared GORM DB handle.
	// Closing the migrate instance would close that handle and break API startup checks.
}
