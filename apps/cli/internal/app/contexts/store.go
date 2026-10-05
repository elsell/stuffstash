package contexts

import "context"

// Update must serialize read-modify-write across processes and write atomically.
type Store interface {
	Load(context.Context) (Config, error)
	Update(context.Context, func(*Config) error) error
}
