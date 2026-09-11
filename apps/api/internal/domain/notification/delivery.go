package notification

import (
	"errors"
	"time"
)

type DeliveryStatus string

const (
	DeliveryPending   DeliveryStatus = "pending"
	DeliveryLeased    DeliveryStatus = "leased"
	DeliveryAccepted  DeliveryStatus = "accepted"
	DeliveryCancelled DeliveryStatus = "cancelled"
	DeliveryFailed    DeliveryStatus = "failed"
)

var (
	ErrInvalidDelivery    = errors.New("invalid notification delivery")
	ErrDeliveryNotDue     = errors.New("notification delivery is not due")
	ErrStaleDeliveryLease = errors.New("notification delivery lease is no longer current")
)

type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaximumDelay time.Duration
}

func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 || p.InitialDelay <= 0 || p.MaximumDelay < p.InitialDelay {
		return ErrInvalidDelivery
	}
	return nil
}
func (p RetryPolicy) delay(attempts int) time.Duration {
	delay := p.InitialDelay
	for step := 1; step < attempts && delay < p.MaximumDelay; step++ {
		if delay > p.MaximumDelay/2 {
			return p.MaximumDelay
		}
		delay *= 2
	}
	return delay
}

// DeliveryState is stored atomically with its lease fence. Repository adapters
// must compare the persisted fence when replacing a claimed state.
type DeliveryState struct {
	Status        DeliveryStatus
	Attempts      int
	NextAttemptAt time.Time
	LeaseUntil    time.Time
	Fence         string
}

func NewDelivery(at time.Time) (DeliveryState, error) {
	if at.IsZero() {
		return DeliveryState{}, ErrInvalidDelivery
	}
	return DeliveryState{Status: DeliveryPending, NextAttemptAt: at}, nil
}
func (d DeliveryState) Claim(now time.Time, fence string, lease time.Duration, policy RetryPolicy) (DeliveryState, error) {
	if now.IsZero() || fence == "" || lease <= 0 || policy.Validate() != nil || d.Attempts < 0 {
		return d, ErrInvalidDelivery
	}
	due := d.Status == DeliveryPending && !d.NextAttemptAt.IsZero() && !now.Before(d.NextAttemptAt)
	expired := d.Status == DeliveryLeased && !d.LeaseUntil.IsZero() && !now.Before(d.LeaseUntil)
	if !due && !expired {
		return d, ErrDeliveryNotDue
	}
	if fence == d.Fence {
		return d, ErrInvalidDelivery
	}
	if d.Attempts >= policy.MaxAttempts {
		return d.terminal(DeliveryFailed), nil
	}
	d.Status = DeliveryLeased
	d.Attempts++
	d.Fence = fence
	d.LeaseUntil = now.Add(lease)
	return d, nil
}
func (d DeliveryState) Accept(now time.Time, fence string) (DeliveryState, error) {
	if !d.owns(now, fence) {
		return d, ErrStaleDeliveryLease
	}
	return d.terminal(DeliveryAccepted), nil
}
func (d DeliveryState) Cancel(now time.Time, fence string) (DeliveryState, error) {
	if !d.owns(now, fence) {
		return d, ErrStaleDeliveryLease
	}
	return d.terminal(DeliveryCancelled), nil
}
func (d DeliveryState) Retry(now time.Time, fence string, policy RetryPolicy) (DeliveryState, error) {
	if !d.owns(now, fence) {
		return d, ErrStaleDeliveryLease
	}
	if policy.Validate() != nil {
		return d, ErrInvalidDelivery
	}
	if d.Attempts >= policy.MaxAttempts {
		return d.terminal(DeliveryFailed), nil
	}
	d.Status = DeliveryPending
	d.NextAttemptAt = now.Add(policy.delay(d.Attempts))
	d.LeaseUntil = time.Time{}
	return d, nil
}
func (d DeliveryState) owns(now time.Time, fence string) bool {
	return !now.IsZero() && d.Status == DeliveryLeased && fence != "" && d.Fence == fence && now.Before(d.LeaseUntil)
}
func (d DeliveryState) terminal(status DeliveryStatus) DeliveryState {
	d.Status = status
	d.LeaseUntil = time.Time{}
	d.NextAttemptAt = time.Time{}
	return d
}
