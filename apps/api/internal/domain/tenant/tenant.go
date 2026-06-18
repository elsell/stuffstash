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

type Tenant struct {
	ID   ID
	Name Name
}
