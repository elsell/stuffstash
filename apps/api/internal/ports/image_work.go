package ports

import "context"

// ImageWorkPriority distinguishes interactive demand from precomputation.
type ImageWorkPriority string

const (
	ImageWorkForeground ImageWorkPriority = "foreground"
	ImageWorkBackground ImageWorkPriority = "background"
)

// ImageWorkAdmission bounds active image work across all callers in a process.
// Acquire waits for capacity or cancellation. The returned release is idempotent.
// Callers must acquire before loading an original and release on every exit.
type ImageWorkAdmission interface {
	Acquire(context.Context, ImageWorkPriority) (release func(), err error)
	// ForegroundWaiting reports queued demand, for cooperative background checkpoints.
	ForegroundWaiting() bool
}
