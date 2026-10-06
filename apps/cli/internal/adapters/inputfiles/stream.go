package inputfiles

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"sync"
)

func (f Files) OpenStream(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if path == "-" {
		if f.StdinTerminal || f.Stdin == nil {
			return nil, ports.Failure("usage", "Pipe the file into stdin, or use a file path instead of -.")
		}
		r, close, err := prepareStdin(f.Stdin)
		if err != nil {
			return nil, ports.Failure("file", "Cannot read stdin. Use a file path instead of -.")
		}
		var once sync.Once
		cleanup := func() { once.Do(close) }
		stop := context.AfterFunc(ctx, cleanup)
		return &streamInput{ctx: ctx, reader: r, close: func() { stop(); cleanup() }}, nil
	}
	file, err := openInput(path)
	if err != nil {
		return nil, ports.Failure("file", "Cannot open the file. Check the path and read permissions.")
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		file.Close()
		return nil, ports.Failure("file", "Choose a readable regular file.")
	}
	stop := context.AfterFunc(ctx, func() { file.Close() })
	return &uploadReader{ctx: ctx, file: file, stop: stop}, nil
}

type streamInput struct {
	ctx    context.Context
	reader io.Reader
	close  func()
}

func (s *streamInput) Read(p []byte) (int, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	return s.reader.Read(p)
}
func (s *streamInput) Close() error { s.close(); return nil }
