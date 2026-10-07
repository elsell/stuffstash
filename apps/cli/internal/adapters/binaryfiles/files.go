package binaryfiles

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"math"
	"os"
)

type Files struct{ Stdout io.Writer }

func (f Files) PublishContent(ctx context.Context, path string, content ports.BinaryContent) error {
	if content.Body == nil {
		return ports.Failure("protocol", "The server returned no file. Try the download again.")
	}
	if path == "" {
		return ports.Failure("usage", "Supply --output PATH or --output -.")
	}
	var out io.Writer = f.Stdout
	var file *os.File
	var finish func() error
	if path != "-" {
		var err error
		file, finish, err = createPrivate(path)
		if err != nil {
			return err
		}
		defer file.Close()
		defer os.Remove(file.Name())
		out = file
	}
	if out == nil {
		return ports.Failure("configuration", "File output is not available. Use --output PATH to save the file.")
	}
	reader := io.Reader(content.Body)
	if content.ContentLength >= 0 && content.ContentLength < math.MaxInt64 {
		reader = io.LimitReader(reader, content.ContentLength+1)
	}
	n, err := io.Copy(out, contextReader{ctx, reader})
	if err != nil {
		return ports.Failure("file", "The CLI could not write the complete file. Examine the connection, disk space, and output permissions. Try again with a new output path.")
	}
	if content.ContentLength >= 0 && n != content.ContentLength {
		return ports.Failure("protocol", "The file size does not match the server response. Try the download again.")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if file != nil {
		if err = file.Sync(); err != nil {
			return ports.Failure("file", "The CLI could not save the file. Examine the available disk space and try again.")
		}
		return finish()
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
