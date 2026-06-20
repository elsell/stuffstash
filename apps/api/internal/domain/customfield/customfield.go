package customfield

import (
	"net/url"
	"regexp"
	"strconv"
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

type AssetTypeID string

func NewAssetTypeID(value string) (AssetTypeID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return AssetTypeID(value), true
}

func (id AssetTypeID) String() string {
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

type Scope string

const (
	ScopeTenant    Scope = "tenant"
	ScopeInventory Scope = "inventory"
)

func (s Scope) String() string {
	return string(s)
}

type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeDate    FieldType = "date"
	FieldTypeURL     FieldType = "url"
	FieldTypeEnum    FieldType = "enum"
)

func NewFieldType(value string) (FieldType, bool) {
	switch FieldType(strings.TrimSpace(value)) {
	case FieldTypeText:
		return FieldTypeText, true
	case FieldTypeNumber:
		return FieldTypeNumber, true
	case FieldTypeBoolean:
		return FieldTypeBoolean, true
	case FieldTypeDate:
		return FieldTypeDate, true
	case FieldTypeURL:
		return FieldTypeURL, true
	case FieldTypeEnum:
		return FieldTypeEnum, true
	default:
		return "", false
	}
}

func (t FieldType) String() string {
	return string(t)
}

type Key string

const maxKeyLength = 80

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,79}$`)

func NewKey(value string) (Key, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxKeyLength || !keyPattern.MatchString(value) {
		return "", false
	}
	return Key(value), true
}

func (k Key) String() string {
	return string(k)
}

type DisplayName string

const maxDisplayNameLength = 120

func NewDisplayName(value string) (DisplayName, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxDisplayNameLength {
		return "", false
	}
	return DisplayName(value), true
}

func (n DisplayName) String() string {
	return string(n)
}

type Description string

const maxDescriptionLength = 1000

func NewDescription(value string) (Description, bool) {
	value = strings.TrimSpace(value)
	if len(value) > maxDescriptionLength {
		return "", false
	}
	return Description(value), true
}

func (d Description) String() string {
	return string(d)
}

type Applicability string

const (
	ApplicabilityAllAssets        Applicability = "all_assets"
	ApplicabilityCustomAssetTypes Applicability = "custom_asset_types"
)

func NewApplicability(value string) (Applicability, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ApplicabilityAllAssets, true
	}
	switch Applicability(value) {
	case ApplicabilityAllAssets:
		return ApplicabilityAllAssets, true
	case ApplicabilityCustomAssetTypes:
		return ApplicabilityCustomAssetTypes, true
	default:
		return "", false
	}
}

func (a Applicability) String() string {
	return string(a)
}

type AssetType struct {
	ID             AssetTypeID
	TenantID       TenantID
	InventoryID    InventoryID
	Scope          Scope
	Key            Key
	DisplayName    DisplayName
	Description    Description
	LifecycleState AssetTypeLifecycleState
}

func NewAssetType(id AssetTypeID, tenantID TenantID, inventoryID InventoryID, scope Scope, key Key, displayName DisplayName, description Description) (AssetType, bool) {
	return NewAssetTypeWithLifecycle(id, tenantID, inventoryID, scope, key, displayName, description, AssetTypeLifecycleActive)
}

type AssetTypeLifecycleState string

const (
	AssetTypeLifecycleActive   AssetTypeLifecycleState = "active"
	AssetTypeLifecycleArchived AssetTypeLifecycleState = "archived"
)

func NewAssetTypeLifecycleState(value string) (AssetTypeLifecycleState, bool) {
	switch AssetTypeLifecycleState(strings.TrimSpace(value)) {
	case AssetTypeLifecycleActive:
		return AssetTypeLifecycleActive, true
	case AssetTypeLifecycleArchived:
		return AssetTypeLifecycleArchived, true
	default:
		return "", false
	}
}

func (s AssetTypeLifecycleState) String() string {
	return string(s)
}

func NewAssetTypeWithLifecycle(id AssetTypeID, tenantID TenantID, inventoryID InventoryID, scope Scope, key Key, displayName DisplayName, description Description, lifecycleState AssetTypeLifecycleState) (AssetType, bool) {
	switch scope {
	case ScopeTenant:
		if inventoryID.String() != "" {
			return AssetType{}, false
		}
	case ScopeInventory:
		if inventoryID.String() == "" {
			return AssetType{}, false
		}
	default:
		return AssetType{}, false
	}
	if _, ok := NewAssetTypeLifecycleState(lifecycleState.String()); !ok {
		return AssetType{}, false
	}
	return AssetType{
		ID:             id,
		TenantID:       tenantID,
		InventoryID:    inventoryID,
		Scope:          scope,
		Key:            key,
		DisplayName:    displayName,
		Description:    description,
		LifecycleState: lifecycleState,
	}, true
}

func (t AssetType) Archive() (AssetType, bool) {
	if t.LifecycleState != AssetTypeLifecycleActive {
		return AssetType{}, false
	}
	t.LifecycleState = AssetTypeLifecycleArchived
	return t, true
}

func (t AssetType) IsActive() bool {
	return t.LifecycleState == AssetTypeLifecycleActive
}

func (t AssetType) CursorKey() string {
	switch t.Scope {
	case ScopeTenant:
		return "0:" + t.ID.String()
	default:
		return "1:" + t.ID.String()
	}
}

func AssetTypesConflict(left AssetType, right AssetType) bool {
	if left.TenantID != right.TenantID || left.Key != right.Key {
		return false
	}
	if left.Scope == ScopeTenant || right.Scope == ScopeTenant {
		return true
	}
	return left.InventoryID == right.InventoryID
}

type Definition struct {
	ID                 ID
	TenantID           TenantID
	InventoryID        InventoryID
	Scope              Scope
	Key                Key
	DisplayName        DisplayName
	Type               FieldType
	EnumOptions        []Key
	Applicability      Applicability
	CustomAssetTypeIDs []AssetTypeID
}

func (d Definition) CursorKey() string {
	switch d.Scope {
	case ScopeTenant:
		return "0:" + d.ID.String()
	default:
		return "1:" + d.ID.String()
	}
}

func DefinitionsConflict(left Definition, right Definition) bool {
	if left.TenantID != right.TenantID || left.Key != right.Key {
		return false
	}
	if left.Scope == ScopeTenant || right.Scope == ScopeTenant {
		return true
	}
	return left.InventoryID == right.InventoryID
}

type DefinitionSet []Definition

func (set DefinitionSet) ValidateValues(values map[string]any) bool {
	return set.ValidateValuesForAssetType(values, "")
}

func (set DefinitionSet) ValidateValuesForAssetType(values map[string]any, customAssetTypeID AssetTypeID) bool {
	byKey := map[Key]Definition{}
	for _, definition := range set {
		if definition.AppliesTo(customAssetTypeID) {
			byKey[definition.Key] = definition
		}
	}

	for rawKey, value := range values {
		key, ok := NewKey(rawKey)
		if !ok {
			return false
		}
		definition, ok := byKey[key]
		if !ok || !definition.ValidValue(value) {
			return false
		}
	}
	return true
}

func NewDefinition(id ID, tenantID TenantID, inventoryID InventoryID, scope Scope, key Key, displayName DisplayName, fieldType FieldType, enumOptions []Key, applicability Applicability, customAssetTypeIDs []AssetTypeID) (Definition, bool) {
	switch scope {
	case ScopeTenant:
		if inventoryID.String() != "" {
			return Definition{}, false
		}
	case ScopeInventory:
		if inventoryID.String() == "" {
			return Definition{}, false
		}
	default:
		return Definition{}, false
	}

	options := append([]Key(nil), enumOptions...)
	if fieldType == FieldTypeEnum {
		if len(options) == 0 || hasDuplicateKeys(options) {
			return Definition{}, false
		}
	} else if len(options) != 0 {
		return Definition{}, false
	}

	targets := append([]AssetTypeID(nil), customAssetTypeIDs...)
	switch applicability {
	case ApplicabilityAllAssets:
		if len(targets) != 0 {
			return Definition{}, false
		}
	case ApplicabilityCustomAssetTypes:
		if len(targets) == 0 || hasDuplicateAssetTypeIDs(targets) {
			return Definition{}, false
		}
	default:
		return Definition{}, false
	}

	return Definition{
		ID:                 id,
		TenantID:           tenantID,
		InventoryID:        inventoryID,
		Scope:              scope,
		Key:                key,
		DisplayName:        displayName,
		Type:               fieldType,
		EnumOptions:        options,
		Applicability:      applicability,
		CustomAssetTypeIDs: targets,
	}, true
}

func (d Definition) AppliesTo(customAssetTypeID AssetTypeID) bool {
	if d.Applicability == ApplicabilityAllAssets {
		return true
	}
	if customAssetTypeID.String() == "" {
		return false
	}
	for _, id := range d.CustomAssetTypeIDs {
		if id == customAssetTypeID {
			return true
		}
	}
	return false
}

func (d Definition) ValidValue(value any) bool {
	switch d.Type {
	case FieldTypeText:
		_, ok := value.(string)
		return ok
	case FieldTypeNumber:
		switch value.(type) {
		case float64, float32, int, int64, int32, jsonNumber:
			return true
		default:
			return false
		}
	case FieldTypeBoolean:
		_, ok := value.(bool)
		return ok
	case FieldTypeDate:
		raw, ok := value.(string)
		if !ok {
			return false
		}
		_, err := time.Parse("2006-01-02", raw)
		return err == nil
	case FieldTypeURL:
		raw, ok := value.(string)
		if !ok {
			return false
		}
		parsed, err := url.Parse(raw)
		return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
	case FieldTypeEnum:
		raw, ok := value.(string)
		if !ok {
			return false
		}
		key, ok := NewKey(raw)
		if !ok {
			return false
		}
		for _, option := range d.EnumOptions {
			if option == key {
				return true
			}
		}
		return false
	default:
		return false
	}
}

type jsonNumber interface {
	Float64() (float64, error)
	Int64() (int64, error)
	String() string
}

func hasDuplicateKeys(keys []Key) bool {
	seen := map[Key]struct{}{}
	for _, key := range keys {
		if _, ok := seen[key]; ok {
			return true
		}
		seen[key] = struct{}{}
	}
	return false
}

func hasDuplicateAssetTypeIDs(ids []AssetTypeID) bool {
	seen := map[AssetTypeID]struct{}{}
	for _, id := range ids {
		if id.String() == "" {
			return true
		}
		if _, exists := seen[id]; exists {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}

func NormalizeJSONNumber(value any) any {
	number, ok := value.(jsonNumber)
	if !ok {
		return value
	}
	raw := number.String()
	if strings.ContainsAny(raw, ".eE") {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			return parsed
		}
	}
	if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return parsed
	}
	return value
}
