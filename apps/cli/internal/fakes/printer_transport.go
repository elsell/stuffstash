package fakes

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

// PrinterTransport simulates a QL-800 protocol peer and physical output. It
// parses the byte stream, begins output on raster data, and emits real status
// frames. Fault settings model hardware conditions rather than call expectations.
type PrinterTransport struct {
	mu                            sync.Mutex
	Connected, Paper, CoverOpen   bool
	DropCompletion                bool
	WaitForStatusDrain            bool
	NotificationsDuringPrint      int
	changed                       chan struct{}
	closed                        bool
	FailAfterBytes                int
	FragmentSize                  int
	PhysicalLabels, StartedLabels int
	Rows                          [][]byte
	input, output                 []byte
	written, rows, expectedRows   int
	printing                      bool
}

func NewPrinterTransport() *PrinterTransport {
	return &PrinterTransport{Connected: true, Paper: true, changed: make(chan struct{})}
}
func (f *PrinterTransport) frame(kind, phase byte) []byte {
	p := make([]byte, 32)
	copy(p, []byte{0x80, 0x20, 0x42, 0x34, 0x38, 0x30, 0x30})
	p[10] = 29
	p[11] = 0x4b
	p[17] = 90
	p[18] = kind
	p[19] = phase
	if !f.Paper {
		p[8] |= 1
	}
	if f.CoverOpen {
		p[9] |= 0x10
	}
	return p
}
func (f *PrinterTransport) notifyLocked() { close(f.changed); f.changed = make(chan struct{}) }
func (f *PrinterTransport) Read(ctx context.Context, p []byte) (int, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		f.mu.Lock()
		if !f.Connected || f.closed {
			f.mu.Unlock()
			return 0, io.EOF
		}
		if len(f.output) > 0 {
			n := len(p)
			if f.FragmentSize > 0 && n > f.FragmentSize {
				n = f.FragmentSize
			}
			if n > len(f.output) {
				n = len(f.output)
			}
			copy(p, f.output[:n])
			f.output = f.output[n:]
			f.notifyLocked()
			f.mu.Unlock()
			return n, nil
		}
		changed := f.changed
		f.mu.Unlock()
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-changed:
		}
	}
}

var errStatusBackpressure = errors.New("printer waiting for status drain")

func (f *PrinterTransport) Write(ctx context.Context, p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.Connected || f.closed {
		return 0, io.ErrClosedPipe
	}
	n := len(p)
	var failure error
	if f.FailAfterBytes > 0 && f.written+n > f.FailAfterBytes {
		n = f.FailAfterBytes - f.written
		if n < 0 {
			n = 0
		}
		failure = io.ErrUnexpectedEOF
	}
	f.written += n
	f.input = append(f.input, p[:n]...)
	for {
		err := f.consume()
		f.notifyLocked()
		if !errors.Is(err, errStatusBackpressure) {
			if err != nil {
				return n, err
			}
			break
		}
		for len(f.output) > 0 {
			changed := f.changed
			f.mu.Unlock()
			select {
			case <-ctx.Done():
				f.mu.Lock()
				return n, ctx.Err()
			case <-changed:
			}
			f.mu.Lock()
			if f.closed || !f.Connected {
				return n, io.ErrClosedPipe
			}
		}
	}
	if failure != nil {
		f.Connected = false
		f.notifyLocked()
	}
	return n, failure
}
func (f *PrinterTransport) consume() error {
	for len(f.input) > 0 {
		data := f.input
		length := 0
		switch data[0] {
		case 0:
			length = 1
		case 0x1b:
			if len(data) < 2 {
				return nil
			}
			if data[1] == 0x40 {
				length = 2
				f.rows = 0
				f.Rows = nil
				f.printing = false
				break
			}
			if len(data) < 3 {
				return nil
			}
			if data[1] != 0x69 {
				return errors.New("invalid ESC command")
			}
			switch data[2] {
			case 0x53:
				length = 3
				if f.printing {
					return errors.New("status request while printing")
				}
				f.output = append(f.output, f.frame(0, 0)...)
			case 0x61, 0x21, 0x4d, 0x41, 0x4b:
				length = 4
			case 0x64:
				length = 5
			case 0x7a:
				length = 13
				if len(data) >= length {
					f.expectedRows = int(binary.LittleEndian.Uint32(data[7:11]))
				}
			default:
				return errors.New("unsupported printer control")
			}
		case 0x67:
			if len(data) < 3 {
				return nil
			}
			length = 3 + int(data[2])
			if len(data) < length {
				return nil
			}
			if data[1] != 0 || data[2] != 90 {
				return errors.New("invalid raster framing")
			}
			if !f.Paper || f.CoverOpen {
				f.output = append(f.output, f.frame(2, 0)...)
				return errors.New("printer needs attention")
			}
			if !f.printing {
				f.printing = true
				f.StartedLabels++
				f.output = append(f.output, f.frame(6, 1)...)
				for i := 0; i < f.NotificationsDuringPrint; i++ {
					f.output = append(f.output, f.frame(5, 1)...)
				}
				if f.WaitForStatusDrain {
					return errStatusBackpressure
				}
			}
			f.Rows = append(f.Rows, append([]byte(nil), data[3:length]...))
			f.rows++
		case 0x1a:
			length = 1
			if f.rows != f.expectedRows || f.rows != 991 {
				return errors.New("incomplete physical label")
			}
			f.PhysicalLabels++
			f.printing = false
			if !f.DropCompletion {
				f.output = append(f.output, f.frame(1, 1)...)
				f.output = append(f.output, f.frame(6, 0)...)
			}
		default:
			return errors.New("unknown printer bytes")
		}
		if len(data) < length {
			return nil
		}
		f.input = f.input[length:]
	}
	return nil
}
func (f *PrinterTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	f.notifyLocked()
	return nil
}
