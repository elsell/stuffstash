package ports

import (
	"context"
	"time"
)

// Delay waits for an operational interval and must return promptly on cancellation.
type Delay interface {
	Wait(context.Context, time.Duration) error
}
