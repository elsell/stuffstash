package ports

import "context"

// InputFiles supplies finite request data without leaking filesystem concerns.
type InputFiles interface {
	Read(context.Context, string) ([]byte, error)
}
