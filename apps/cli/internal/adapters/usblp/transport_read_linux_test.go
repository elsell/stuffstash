//go:build linux

package usblp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// A datagram peer models completed USB IN transfers: an empty transfer makes
// read return (0,nil), while later status fragments remain available. The real
// kernel supplies nonblocking reads and poll readiness; no syscalls are mocked.
func usbReadPeer(t *testing.T) (*transport, int) {
	t.Helper()
	pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	device := &transport{fd: pair[0]}
	t.Cleanup(func() { _ = device.Close(); _ = unix.Close(pair[1]) })
	return device, pair[1]
}
func sendTransfer(t *testing.T, peer int, data []byte) {
	t.Helper()
	if _, err := unix.SendmsgN(peer, data, nil, nil, 0); err != nil {
		t.Fatal(err)
	}
}
func TestReadPreservesStatusAcrossEmptyUSBCompletions(t *testing.T) {
	device, peer := usbReadPeer(t)
	frame := make([]byte, 32)
	copy(frame, []byte{0x80, 0x20, 0x42, 0x34, 0x38})
	frame[18] = 1
	sendTransfer(t, peer, nil)
	sendTransfer(t, peer, frame[:7])
	sendTransfer(t, peer, nil)
	sendTransfer(t, peer, frame[7:])
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	received := make([]byte, len(frame))
	offset := 0
	for offset < len(received) {
		n, err := device.Read(ctx, received[offset:])
		if err != nil {
			t.Fatalf("healthy USB transfer treated as failure after %d bytes: %v", offset, err)
		}
		if n == 0 {
			t.Fatal("empty USB completion escaped transport")
		}
		offset += n
	}
	if !bytes.Equal(received, frame) {
		t.Fatalf("status bytes changed: %x", received)
	}
}
func TestReadAfterEmptyUSBCompletionRemainsCancellable(t *testing.T) {
	device, peer := usbReadPeer(t)
	sendTransfer(t, peer, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	n, err := device.Read(ctx, make([]byte, 32))
	if n != 0 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("empty completion became disconnect or ignored deadline: %d %v", n, err)
	}
}
func TestReadStillReportsClosedDescriptor(t *testing.T) {
	device, _ := usbReadPeer(t)
	if err := device.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := device.Read(ctx, make([]byte, 32)); !errors.Is(err, unix.EBADF) {
		t.Fatalf("device error was hidden: %v", err)
	}
}

func TestReadDoesNotIgnoreHangupAfterEmptyRead(t *testing.T) {
	pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	device := &transport{fd: pair[0]}
	defer device.Close()
	if err := unix.Close(pair[1]); err != nil {
		t.Fatal(err)
	}
	// A closed stream supplies read(0,nil) and POLLHUP. Unlike a healthy USB
	// zero-length completion, that poll event must terminate the operation.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := device.Read(ctx, make([]byte, 32)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("hangup was not retained: %v", err)
	}
}

func TestRepeatedEmptyCompletionsDoNotBusyLoop(t *testing.T) {
	device, peer := usbReadPeer(t)
	for i := 0; i < 10; i++ {
		sendTransfer(t, peer, nil)
	}
	sendTransfer(t, peer, []byte{0x80})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	n, err := device.Read(ctx, make([]byte, 32))
	if n != 0 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("empty completion burst was consumed without backoff: %d %v", n, err)
	}
	// Deadline cancellation must leave the remaining transfers available.
	next, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	data := make([]byte, 32)
	n, err = device.Read(next, data)
	if err != nil || n != 1 || data[0] != 0x80 {
		t.Fatalf("later status lost: %d %v", n, err)
	}
}
