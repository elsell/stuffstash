// Package printprocess provides process-local timing and cryptographic attempt
// identities. None of these values confer authorization without the API claim.
package printprocess

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"time"
)

type Identities struct{}

func (Identities) Attempt() (string, error)    { return secret() }
func (Identities) ClaimToken() (string, error) { return secret() }
func secret() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

type Waiter struct{}

func (Waiter) Wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type Backoff struct{ Minimum, Maximum time.Duration }

func (b Backoff) Delay(failures int) time.Duration {
	delay := b.Minimum
	for step := 1; step < failures && delay < b.Maximum; step++ {
		if delay > b.Maximum/2 {
			delay = b.Maximum
			break
		}
		delay *= 2
	}
	if delay > b.Maximum {
		delay = b.Maximum
	}
	// Equal jitter bounds reconnection load without an immediate retry burst.
	half := delay / 2
	jitter, err := rand.Int(rand.Reader, big.NewInt(int64(delay-half)+1))
	if err != nil {
		return delay
	}
	return half + time.Duration(jitter.Int64())
}
