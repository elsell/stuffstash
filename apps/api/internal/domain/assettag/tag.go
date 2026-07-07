package assettag

import (
	"regexp"
	"strings"
	"time"
)

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

type Key string

const maxKeyLength = 80

var keyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,79}$`)

func NewKey(value string) (Key, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxKeyLength || !keyPattern.MatchString(value) {
		return "", false
	}
	return Key(value), true
}

func KeyFromDisplayName(value string) (Key, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastHyphen := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastHyphen = false
		default:
			if builder.Len() > 0 && !lastHyphen {
				builder.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	key := strings.Trim(builder.String(), "-")
	if len(key) > maxKeyLength {
		key = strings.TrimRight(key[:maxKeyLength], "-")
	}
	return NewKey(key)
}

func (k Key) String() string {
	return string(k)
}

type DisplayName string

func NewDisplayName(value string) (DisplayName, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 80 {
		return "", false
	}
	return DisplayName(value), true
}

func (n DisplayName) String() string {
	return string(n)
}

type Color string

func NewColor(value string) (Color, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", true
	}
	if !strings.HasPrefix(value, "#") {
		value = "#" + value
	}
	if len(value) != 7 {
		return "", false
	}
	for _, r := range value[1:] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return "", false
		}
	}
	return Color(strings.ToUpper(value)), true
}

func (c Color) String() string {
	return string(c)
}

type LifecycleState string

const (
	LifecycleStateActive   LifecycleState = "active"
	LifecycleStateArchived LifecycleState = "archived"
)

func (s LifecycleState) String() string {
	return string(s)
}

type Tag struct {
	ID             ID
	TenantID       TenantID
	InventoryID    InventoryID
	Key            Key
	DisplayName    DisplayName
	Color          Color
	LifecycleState LifecycleState
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewTag(id ID, tenantID TenantID, inventoryID InventoryID, key Key, displayName DisplayName, color Color, now time.Time) (Tag, bool) {
	if id.String() == "" || tenantID.String() == "" || inventoryID.String() == "" || key.String() == "" || displayName.String() == "" || now.IsZero() {
		return Tag{}, false
	}
	return Tag{
		ID:             id,
		TenantID:       tenantID,
		InventoryID:    inventoryID,
		Key:            key,
		DisplayName:    displayName,
		Color:          color,
		LifecycleState: LifecycleStateActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, true
}
