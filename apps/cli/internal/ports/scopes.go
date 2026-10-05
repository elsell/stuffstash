package ports

import "context"

type Access struct {
	Relationship string   `json:"relationship"`
	Permissions  []string `json:"permissions"`
}

type Tenant struct {
	Access    Access `json:"access"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Lifecycle string `json:"lifecycleState"`
}
type ScopeCatalog interface {
	Tenants(context.Context, Page) (Result[[]Tenant], error)
	Inventories(context.Context, Scope, Page) (Result[[]Inventory], error)
}
type Choice struct{ ID, Label, Detail string }
type Selector interface {
	Pick(context.Context, string, []Choice) (string, error)
}
