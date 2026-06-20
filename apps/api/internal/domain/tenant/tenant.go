package tenant

import "strings"

type ID string

func NewID(value string) (ID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return ID(value), true
}

func (id ID) String() string {
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

type Tenant struct {
	ID             ID
	Name           Name
	LifecycleState LifecycleState
}

func NewTenant(id ID, name Name, lifecycleState LifecycleState) (Tenant, bool) {
	if id.String() == "" || name.String() == "" {
		return Tenant{}, false
	}
	if lifecycleState.String() == "" {
		lifecycleState = LifecycleStateActive
	}
	if _, ok := NewLifecycleState(lifecycleState.String()); !ok {
		return Tenant{}, false
	}
	return Tenant{ID: id, Name: name, LifecycleState: lifecycleState}, true
}

func (t Tenant) IsActive() bool {
	return t.LifecycleState == "" || t.LifecycleState == LifecycleStateActive
}
