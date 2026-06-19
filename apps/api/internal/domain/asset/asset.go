package asset

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

type TenantID string

func (id TenantID) String() string {
	return string(id)
}

type InventoryID string

func (id InventoryID) String() string {
	return string(id)
}

type Kind string

const (
	KindItem      Kind = "item"
	KindContainer Kind = "container"
	KindLocation  Kind = "location"
)

func NewKind(value string) (Kind, bool) {
	switch Kind(strings.TrimSpace(value)) {
	case KindItem:
		return KindItem, true
	case KindContainer:
		return KindContainer, true
	case KindLocation:
		return KindLocation, true
	default:
		return "", false
	}
}

func (k Kind) String() string {
	return string(k)
}

func (k Kind) CanContainChildren() bool {
	return k == KindContainer || k == KindLocation
}

type LifecycleState string

const (
	LifecycleStateActive   LifecycleState = "active"
	LifecycleStateArchived LifecycleState = "archived"
)

func (s LifecycleState) String() string {
	return string(s)
}

type Title string

const maxTitleLength = 160

func NewTitle(value string) (Title, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxTitleLength {
		return "", false
	}
	return Title(value), true
}

func (t Title) String() string {
	return string(t)
}

type Description string

func NewDescription(value string) Description {
	return Description(strings.TrimSpace(value))
}

func (d Description) String() string {
	return string(d)
}

type CustomFields struct {
	values map[string]any
}

func NewCustomFields(values map[string]any) (CustomFields, bool) {
	copied := map[string]any{}
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			return CustomFields{}, false
		}
		copied[key] = value
	}
	return CustomFields{values: copied}, true
}

func NewEmptyCustomFields() CustomFields {
	return CustomFields{values: map[string]any{}}
}

func (fields CustomFields) Values() map[string]any {
	copied := map[string]any{}
	for key, value := range fields.values {
		copied[key] = value
	}
	return copied
}

func (fields CustomFields) IsEmpty() bool {
	return len(fields.values) == 0
}

func NewEmptyOnlyCustomFields(values map[string]any) (CustomFields, bool) {
	if len(values) != 0 {
		return CustomFields{}, false
	}
	return NewEmptyCustomFields(), true
}

type Asset struct {
	ID             ID
	TenantID       TenantID
	InventoryID    InventoryID
	ParentAssetID  ID
	Kind           Kind
	Title          Title
	Description    Description
	CustomFields   CustomFields
	LifecycleState LifecycleState
}
