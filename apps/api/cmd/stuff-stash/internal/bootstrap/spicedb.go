package bootstrap

import (
	"context"
	"fmt"
	"os"
	"time"
)

type schemaBootstrapper interface {
	BootstrapSchema(ctx context.Context, schema string) error
}

func bootstrapSpiceDBSchema(ctx context.Context, authorizer schemaBootstrapper, schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()

	var lastErr error
	for {
		if err := authorizer.BootstrapSchema(ctx, string(schema)); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("bootstrap spicedb schema: %w", lastErr)
		case <-ticker.C:
		}
	}
}
