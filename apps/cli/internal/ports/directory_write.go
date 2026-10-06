package ports

import "context"

type DirectoryWrite string

const (
	CreateTenant    DirectoryWrite = "create-tenant"
	UpdateTenant    DirectoryWrite = "update-tenant"
	CreateInventory DirectoryWrite = "create-inventory"
	UpdateInventory DirectoryWrite = "update-inventory"
)

type DirectoryWriter interface {
	WriteDirectory(context.Context, DirectoryWrite, Scope, []byte) (any, error)
}
