package ports

import "context"

type Principal struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"displayName,omitempty"`
	Email       *string `json:"email,omitempty"`
}

type Directory interface {
	ScopeCatalog
	Principal(context.Context) (Result[Principal], error)
	Tenant(context.Context, Scope) (Result[Tenant], error)
	Inventory(context.Context, Scope) (Result[Inventory], error)
}
