package ports

import "context"

type SecretInput interface {
	ReadSecret(context.Context, string, int) (string, error)
}
