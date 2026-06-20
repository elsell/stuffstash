package inventory

import "strings"

type TenantID string

func (id TenantID) String() string {
	return string(id)
}

type InventoryID string

func NewID(value string) (InventoryID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return InventoryID(value), true
}

func (id InventoryID) String() string {
	return string(id)
}

type Name string

func NewName(value string) (Name, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return Name(value), true
}

func (n Name) String() string {
	return string(n)
}

type LifecycleState string

const (
	LifecycleStateActive   LifecycleState = "active"
	LifecycleStateArchived LifecycleState = "archived"
)

func NewLifecycleState(value string) (LifecycleState, bool) {
	switch LifecycleState(strings.TrimSpace(value)) {
	case LifecycleStateActive:
		return LifecycleStateActive, true
	case LifecycleStateArchived:
		return LifecycleStateArchived, true
	default:
		return "", false
	}
}

func (s LifecycleState) String() string {
	return string(s)
}

type Inventory struct {
	ID             InventoryID
	TenantID       TenantID
	Name           Name
	LifecycleState LifecycleState
}

func NewInventory(id InventoryID, tenantID TenantID, name Name, lifecycleState LifecycleState) (Inventory, bool) {
	if id.String() == "" || tenantID.String() == "" || name.String() == "" {
		return Inventory{}, false
	}
	if lifecycleState.String() == "" {
		lifecycleState = LifecycleStateActive
	}
	if _, ok := NewLifecycleState(lifecycleState.String()); !ok {
		return Inventory{}, false
	}
	return Inventory{ID: id, TenantID: tenantID, Name: name, LifecycleState: lifecycleState}, true
}

func (i Inventory) IsActive() bool {
	return i.LifecycleState == "" || i.LifecycleState == LifecycleStateActive
}
