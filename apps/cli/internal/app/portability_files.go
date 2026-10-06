package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
)

const maximumImportCSVBytes = 10 << 20
const maximumImportJSONBytes = 16 << 20

func (r Runner) readImportFile(ctx context.Context, path string, limit int64) ([]byte, error) {
	if r.StreamFiles == nil {
		return nil, ports.Failure("configuration", "Streaming input is unavailable. Update the CLI.")
	}
	body, err := r.StreamFiles.OpenStream(ctx, path)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, ports.Failure("input", "Cannot read import input. Check the file or pipe and try again.")
	}
	if int64(len(data)) > limit {
		return nil, ports.Failure("input", "Import input exceeds its supported limit: 10 MiB for CSV, 16 MiB for JSON.")
	}
	return data, nil
}
