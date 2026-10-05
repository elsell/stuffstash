package ports

import "context"

type TextInput interface {
	ReadText(context.Context, string, int) (string, error)
}
