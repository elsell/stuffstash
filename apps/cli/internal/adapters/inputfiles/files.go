package inputfiles

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"os"
)

const maximumBytes = 1 << 20

type Files struct {
	Stdin         io.Reader
	StdinTerminal bool
}

func (f Files) Read(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var reader io.Reader
	if path == "-" {
		if f.StdinTerminal || f.Stdin == nil {
			return nil, ports.Failure("usage", "Pipe JSON into stdin or use --input FILE.")
		}
		var cleanup func()
		var err error
		reader, cleanup, err = prepareStdin(f.Stdin)
		if err != nil {
			return nil, ports.Failure("input", "The CLI cannot open stdin. Use --input FILE instead.")
		}
		defer cleanup()
	} else {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			return nil, ports.Failure("input", "The CLI cannot read the input file. Use a readable regular file.")
		}
		file, err := openInput(path)
		if err != nil {
			return nil, ports.Failure("input", "The CLI cannot open the input file. Examine the path and file permissions.")
		}
		defer file.Close()
		info, err = file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			return nil, ports.Failure("input", "The input path is not a regular file. Select another file.")
		}
		reader = file
	}
	// Closing an input pipe on process cancellation releases a blocked read.
	// Finite commands own stdin for their lifetime; do not leave a read goroutine.
	if closer, ok := reader.(io.Closer); ok {
		stop := context.AfterFunc(ctx, func() { _ = closer.Close() })
		defer stop()
	}
	data, err := io.ReadAll(io.LimitReader(reader, maximumBytes+1))
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, ports.Failure("input", "The CLI cannot read the input. Examine the file or pipe and try again.")
	}
	if len(data) > maximumBytes {
		return nil, ports.Failure("input", "The input exceeds 1 MiB. Use a smaller JSON request.")
	}
	return data, nil
}
