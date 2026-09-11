package notification

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestDeliveryFencesExpiredWorkersAndReclaimsLease(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	policy := RetryPolicy{MaxAttempts: 3, InitialDelay: time.Second, MaximumDelay: time.Minute}
	pending, err := NewDelivery(now)
	if err != nil {
		t.Fatal(err)
	}
	first, err := pending.Claim(now, "first", time.Minute, policy)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Accept(now, "wrong"); !errors.Is(err, ErrStaleDeliveryLease) {
		t.Fatalf("wrong worker: %v", err)
	}
	later := now.Add(time.Minute)
	if _, err := first.Accept(later, "first"); !errors.Is(err, ErrStaleDeliveryLease) {
		t.Fatalf("expired worker: %v", err)
	}
	second, err := first.Claim(later, "second", time.Minute, policy)
	if err != nil {
		t.Fatal(err)
	}
	if second.Attempts != 2 {
		t.Fatalf("attempts %d", second.Attempts)
	}
	if _, err := second.Accept(later, "first"); !errors.Is(err, ErrStaleDeliveryLease) {
		t.Fatalf("reclaimed fence: %v", err)
	}
	accepted, err := second.Accept(later, "second")
	if err != nil || accepted.Status != DeliveryAccepted {
		t.Fatalf("accept: %+v %v", accepted, err)
	}
	if _, err := accepted.Claim(later, "third", time.Minute, policy); !errors.Is(err, ErrDeliveryNotDue) {
		t.Fatalf("terminal claim: %v", err)
	}
}

func TestDeliveryRetryBackoffAndExhaustion(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	policy := RetryPolicy{MaxAttempts: 3, InitialDelay: time.Second, MaximumDelay: 2 * time.Second}
	value, _ := NewDelivery(now)
	for attempt := 1; attempt <= 3; attempt++ {
		fence := "worker-" + strconv.Itoa(attempt)
		claimed, err := value.Claim(now, fence, time.Minute, policy)
		if err != nil {
			t.Fatal(err)
		}
		value, err = claimed.Retry(now, fence, policy)
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 3 {
			if value.Status != DeliveryFailed {
				t.Fatalf("not exhausted: %+v", value)
			}
		} else {
			delay := time.Duration(attempt) * time.Second
			if value.Status != DeliveryPending || !value.NextAttemptAt.Equal(now.Add(delay)) {
				t.Fatalf("backoff: %+v", value)
			}
			if _, err := value.Claim(now, "early", time.Minute, policy); !errors.Is(err, ErrDeliveryNotDue) {
				t.Fatalf("early claim: %v", err)
			}
			now = value.NextAttemptAt
		}
	}
}

func TestDeliveryCancellationAndInvalidPolicy(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	value, _ := NewDelivery(now)
	if _, err := value.Claim(now, "worker", time.Minute, RetryPolicy{}); !errors.Is(err, ErrInvalidDelivery) {
		t.Fatal(err)
	}
	policy := RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
	claimed, _ := value.Claim(now, "worker", time.Minute, policy)
	cancelled, err := claimed.Cancel(now, "worker")
	if err != nil || cancelled.Status != DeliveryCancelled {
		t.Fatalf("cancel: %+v %v", cancelled, err)
	}
	if _, err := cancelled.Retry(now, "worker", policy); !errors.Is(err, ErrStaleDeliveryLease) {
		t.Fatalf("cancel retry: %v", err)
	}
}

func TestExpiredFinalAttemptBecomesTerminalAndCannotReuseFence(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	policy := RetryPolicy{MaxAttempts: 1, InitialDelay: time.Second, MaximumDelay: time.Minute}
	pending, _ := NewDelivery(now)
	claimed, _ := pending.Claim(now, "first", time.Second, policy)
	if _, err := claimed.Claim(now.Add(time.Second), "first", time.Second, policy); !errors.Is(err, ErrInvalidDelivery) {
		t.Fatalf("reused fence: %v", err)
	}
	exhausted, err := claimed.Claim(now.Add(time.Second), "replacement", time.Second, policy)
	if err != nil || exhausted.Status != DeliveryFailed || exhausted.Attempts != 1 {
		t.Fatalf("exhaustion: %+v %v", exhausted, err)
	}
}

func TestRetryDelayCannotOverflow(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	maximum := time.Duration(1<<63 - 1)
	policy := RetryPolicy{MaxAttempts: 100, InitialDelay: maximum/2 + 1, MaximumDelay: maximum}
	state := DeliveryState{Status: DeliveryLeased, Attempts: 99, Fence: "worker", LeaseUntil: now.Add(time.Minute)}
	next, err := state.Retry(now, "worker", policy)
	if err != nil || !next.NextAttemptAt.Equal(now.Add(maximum)) {
		t.Fatalf("overflow retry: %+v %v", next, err)
	}
}
