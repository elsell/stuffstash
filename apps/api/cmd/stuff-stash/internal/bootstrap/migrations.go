package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/dbmigrations"
	"github.com/stuffstash/stuff-stash/internal/config"
	"gorm.io/gorm"
)

func RunMigrationCommand(ctx context.Context, cfg config.Config, args []string, output io.Writer) error {
	if len(args) != 1 {
		return errors.New("migration command must be one of: up, status")
	}
	switch args[0] {
	case "up", "status":
	default:
		return errors.New("migration command must be one of: up, status")
	}
	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return errors.New("database dsn is required")
	}

	db, closeDB, err := openPostgresDB(ctx, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer func() {
		_ = closeDB()
	}()

	runner := dbmigrations.NewRunner(db)
	switch args[0] {
	case "up":
		if err := runner.Up(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(output, "migrations applied")
	case "status":
		status, err := runner.Status()
		if err != nil {
			return err
		}
		if status.Empty {
			_, _ = fmt.Fprintln(output, "migration version: none")
			return nil
		}
		_, _ = fmt.Fprintf(output, "migration version: %d latest: %d dirty: %t\n", status.Version, status.Latest, status.Dirty)
	}
	return nil
}

func verifyPostgresSchemaCurrent(db *gorm.DB) error {
	status, err := dbmigrations.NewRunner(db).Status()
	if err != nil {
		return err
	}
	return validateMigrationStatus(status)
}

func validateMigrationStatus(status dbmigrations.Status) error {
	if status.Empty {
		return errors.New("database migrations have not been applied")
	}
	if status.Dirty {
		return fmt.Errorf("database migrations are dirty at version %d", status.Version)
	}
	if status.Version != status.Latest {
		return fmt.Errorf("database migration version %d does not match latest %d", status.Version, status.Latest)
	}
	return nil
}
