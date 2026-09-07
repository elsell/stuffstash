package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
	"net/url"
	"strconv"
	"strings"
)

func RunThumbnailJobsCommand(ctx context.Context, cfg config.Config, args []string, output io.Writer, observer ports.Observer) error {
	if len(args) == 0 {
		return errors.New("thumbnail-jobs requires status or retry-failed")
	}
	limit := 100
	switch args[0] {
	case "status":
		if len(args) != 1 {
			return errors.New("status does not accept arguments")
		}
	case "retry-failed":
		flags := flag.NewFlagSet("retry-failed", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		flags.IntVar(&limit, "limit", 100, "maximum failed jobs to retry")
		if err := flags.Parse(args[1:]); err != nil {
			return errors.New("invalid retry-failed arguments")
		}
		if flags.NArg() != 0 || limit < 1 || limit > 1000 {
			return errors.New("retry limit must be between 1 and 1000")
		}
	default:
		return errors.New("thumbnail-jobs requires status or retry-failed")
	}
	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return errors.New("database dsn is required")
	}
	var repository ports.ThumbnailJobOperations
	var closeStore func() error
	var err error
	switch strings.ToLower(strings.TrimSpace(cfg.RepositoryMode)) {
	case "postgres":
		repository, closeStore, err = openPostgresStore(ctx, cfg.DatabaseDSN)
	case "sqlite":
		repository, closeStore, err = openThumbnailSQLiteStore(ctx, cfg.DatabaseDSN)
	default:
		return errors.New("thumbnail-jobs requires postgres or sqlite storage")
	}
	if err != nil {
		return err
	}
	defer func() { _ = closeStore() }()
	clock := ports.SystemClock{}
	now := clock.Now()
	if args[0] == "retry-failed" {
		count, err := repository.RetryFailedThumbnailJobs(ctx, limit, now)
		if err != nil {
			return err
		}
		if observer != nil {
			observer.Record(ctx, ports.Event{Name: ports.EventThumbnailJobsRetried, Message: "failed thumbnail jobs retried", Fields: map[string]string{"count": strconv.Itoa(count)}})
		}
		return json.NewEncoder(output).Encode(map[string]any{"retried": count})
	}
	status, err := repository.ThumbnailQueueStatus(ctx, now)
	if err != nil {
		return err
	}
	age := float64(0)
	if !status.OldestPendingAt.IsZero() {
		age = now.Sub(status.OldestPendingAt).Seconds()
		if age < 0 {
			age = 0
		}
	}
	return json.NewEncoder(output).Encode(map[string]any{"pending": status.Pending, "leased": status.Leased, "failed": status.Failed, "completed": status.Completed, "oldest_pending_age_seconds": age, "backfill_complete": status.BackfillComplete})
}

func openThumbnailSQLiteStore(ctx context.Context, dsn string) (ports.ThumbnailJobOperations, func() error, error) {
	path, queryText, _ := strings.Cut(strings.TrimPrefix(strings.TrimSpace(dsn), "file:"), "?")
	if path == "" || path == ":memory:" {
		return nil, nil, errors.New("thumbnail-jobs requires an existing SQLite file")
	}
	query, err := url.ParseQuery(queryText)
	if err != nil {
		return nil, nil, errors.New("invalid SQLite configuration")
	}
	query.Set("mode", "rw")
	db, err := gormstore.OpenSQLite("file:" + path + "?" + query.Encode())
	if err != nil {
		return nil, nil, err
	}
	pool, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	if err := pool.PingContext(ctx); err != nil {
		_ = pool.Close()
		return nil, nil, err
	}
	return gormstore.NewStore(db), pool.Close, nil
}
