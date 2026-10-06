package ports

import "context"

type LifecycleAction string

const (
	Archive LifecycleAction = "archive"
	Restore LifecycleAction = "restore"
	Delete  LifecycleAction = "delete"
)

type DirectoryResource string

const (
	HouseholdResource DirectoryResource = "household"
	InventoryResource DirectoryResource = "inventory"
)

type DirectoryLifecycle interface {
	ChangeDirectoryLifecycle(context.Context, DirectoryResource, LifecycleAction, Scope) (any, error)
}
