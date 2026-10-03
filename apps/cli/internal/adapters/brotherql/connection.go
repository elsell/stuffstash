package brotherql

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
)

type Connection struct {
	finished                   *printing.Submission
	transport                  ports.PrinterTransport
	active                     *printing.Submission
	ready, complete, uncertain bool
}

func NewConnection(transport ports.PrinterTransport) *Connection {
	return &Connection{transport: transport}
}
func (c *Connection) Close() error { return c.transport.Close() }
func (c *Connection) Readiness(ctx context.Context) (printing.Readiness, error) {
	if c.active != nil || c.uncertain {
		return printing.Readiness{State: printing.Busy, Reason: printing.ActiveSubmission}, nil
	}
	c.ready = false
	if _, err := writeAll(ctx, c.transport, []byte{0x1b, 0x69, 0x53}); err != nil {
		return printing.Readiness{State: printing.Unavailable, Reason: printing.Disconnected}, nil
	}
	for i := 0; i < 32; i++ {
		s, err := c.readStatus(ctx)
		if err != nil {
			return printing.Readiness{State: printing.Unknown, Reason: printing.InvalidStatus}, err
		}
		if s.reason != printing.NoReason {
			return printing.Readiness{State: printing.NeedsAttention, Reason: s.reason}, nil
		}
		if s.kind != statusReply {
			continue
		}
		if s.phase != 0 {
			return printing.Readiness{State: printing.Busy}, nil
		}
		c.ready = true
		return printing.Readiness{State: printing.Ready}, nil
	}
	return printing.Readiness{State: printing.Unknown, Reason: printing.InvalidStatus}, errors.New("no status reply received")
}
func (c *Connection) Submit(ctx context.Context, label printing.Label) (printing.Submission, error) {
	empty := printing.Submission{}
	if c.uncertain {
		return empty, &printing.SubmissionError{Outcome: printing.Uncertain, Reason: printing.ActiveSubmission}
	}
	if c.active != nil || !c.ready {
		return empty, &printing.SubmissionError{Outcome: printing.NoOutput, Reason: printing.ActiveSubmission}
	}
	data, err := encode(label)
	if err != nil {
		return empty, &printing.SubmissionError{Outcome: printing.NoOutput, Reason: printing.InvalidArtifact}
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return empty, &printing.SubmissionError{Outcome: printing.NoOutput, Reason: printing.DeviceError}
	}
	submission := printing.Submission{Reference: hex.EncodeToString(random[:]), AttemptID: label.AttemptID, Copy: label.Copy}
	c.ready = false
	c.complete = false
	sent, err := writeAll(ctx, c.transport, data)
	if err != nil {
		outcome := printing.NoOutput
		if sent > 0 {
			outcome = printing.Uncertain
			c.uncertain = true
		}
		return empty, &printing.SubmissionError{Outcome: outcome, Reason: printing.Interrupted}
	}
	c.active = &submission
	return submission, nil
}
func (c *Connection) Observe(ctx context.Context, submission printing.Submission) (printing.Observation, error) {
	if c.finished != nil && *c.finished == submission {
		return printing.Observation{Outcome: printing.Completed}, nil
	}
	if c.active == nil || *c.active != submission {
		return printing.Observation{Outcome: printing.Uncertain, Reason: printing.UnrecognizedSubmission}, nil
	}
	if c.uncertain {
		return printing.Observation{Outcome: printing.Uncertain, Reason: printing.Interrupted}, nil
	}
	s, err := c.readStatus(ctx)
	if err != nil || s.reason != printing.NoReason {
		c.uncertain = true
		reason := s.reason
		if err != nil {
			reason = printing.Interrupted
		}
		return printing.Observation{Outcome: printing.Uncertain, Reason: reason}, nil
	}
	if s.kind == statusComplete {
		c.complete = true
	}
	if c.complete && s.kind == statusPhase && s.phase == 0 {
		c.finished = c.active
		c.active = nil
		c.complete = false
		return printing.Observation{Outcome: printing.Completed}, nil
	}
	return printing.Observation{Outcome: printing.Pending}, nil
}
func (c *Connection) readStatus(ctx context.Context) (status, error) {
	var packet [32]byte
	read := 0
	for read < len(packet) {
		n, err := c.transport.Read(ctx, packet[read:])
		read += n
		if err != nil {
			return status{}, err
		}
		if n == 0 {
			return status{}, io.ErrNoProgress
		}
	}
	return parseStatus(packet[:])
}
func writeAll(ctx context.Context, transport ports.PrinterTransport, data []byte) (int, error) {
	written := 0
	for written < len(data) {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		n, err := transport.Write(ctx, data[written:])
		written += n
		if err != nil {
			return written, err
		}
		if n == 0 {
			return written, io.ErrNoProgress
		}
	}
	return written, nil
}

// ConfirmIdle requires a fresh status response and refuses local submission
// ambiguity. Recovery opens a new locked connection before calling this method.
func (c *Connection) ConfirmIdle(ctx context.Context) error {
	state, err := c.Readiness(ctx)
	if err != nil {
		return err
	}
	if state.State != printing.Ready {
		return ports.ErrRecoveryRequired
	}
	return nil
}
