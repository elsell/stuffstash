package ports

import "context"

type EventName string

const (
	EventHealthChecked EventName = "health.checked"
)

type Event struct {
	Name    EventName
	Message string
	Fields  map[string]string
}

type Observer interface {
	Record(ctx context.Context, event Event)
}
