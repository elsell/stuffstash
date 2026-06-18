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

type Inventory struct {
	ID       InventoryID
	TenantID TenantID
	Name     Name
}
