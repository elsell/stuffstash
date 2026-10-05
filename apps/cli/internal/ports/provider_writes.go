package ports

import "context"

type ProviderWrites interface {
	CreateProvider(context.Context, string, []byte) (Result[ProviderProfile], error)
	UpdateProvider(context.Context, string, string, []byte) (Result[ProviderProfile], error)
	ReplaceProviderCredential(context.Context, string, string, []byte) (Result[ProviderProfile], error)
}
