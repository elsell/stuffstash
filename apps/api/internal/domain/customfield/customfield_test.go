package customfield

import "testing"

func TestDefinitionSetValidatesValuesByType(t *testing.T) {
	serial := definition(t, "serial", FieldTypeText, nil)
	count := definition(t, "count", FieldTypeNumber, nil)
	owned := definition(t, "owned", FieldTypeBoolean, nil)
	expiry := definition(t, "expires-on", FieldTypeDate, nil)
	link := definition(t, "manual-url", FieldTypeURL, nil)
	status := definition(t, "status", FieldTypeEnum, []string{"new", "used"})

	values := map[string]any{
		"serial":     "abc",
		"count":      float64(2),
		"owned":      true,
		"expires-on": "2026-06-19",
		"manual-url": "https://stuffstash.online/manual",
		"status":     "used",
	}
	if !(DefinitionSet{serial, count, owned, expiry, link, status}).ValidateValues(values) {
		t.Fatalf("expected valid custom field values")
	}
}

func TestDefinitionSetRejectsUnknownAndInvalidValues(t *testing.T) {
	definitions := DefinitionSet{
		definition(t, "serial", FieldTypeText, nil),
		definition(t, "count", FieldTypeNumber, nil),
		definition(t, "expires-on", FieldTypeDate, nil),
		definition(t, "manual-url", FieldTypeURL, nil),
		definition(t, "status", FieldTypeEnum, []string{"new", "used"}),
	}

	cases := []struct {
		name   string
		values map[string]any
	}{
		{name: "unknown field", values: map[string]any{"hidden": "value"}},
		{name: "bad key", values: map[string]any{"Bad": "value"}},
		{name: "wrong text type", values: map[string]any{"serial": true}},
		{name: "wrong number type", values: map[string]any{"count": "2"}},
		{name: "bad date", values: map[string]any{"expires-on": "06/19/2026"}},
		{name: "bad url", values: map[string]any{"manual-url": "ftp://example.com"}},
		{name: "bad enum", values: map[string]any{"status": "broken"}},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if definitions.ValidateValues(item.values) {
				t.Fatalf("expected values to be rejected: %+v", item.values)
			}
		})
	}
}

func TestDefinitionSetValidatesValuesByCustomAssetType(t *testing.T) {
	serial := definition(t, "serial", FieldTypeText, nil)
	dose := targetedDefinition(t, "dose", FieldTypeText, "medicine-type")

	if !(DefinitionSet{serial, dose}).ValidateValuesForAssetType(map[string]any{
		"serial": "abc",
		"dose":   "20mg",
	}, "medicine-type") {
		t.Fatalf("expected field targeted to matching custom asset type to be valid")
	}
	if (DefinitionSet{serial, dose}).ValidateValuesForAssetType(map[string]any{
		"dose": "20mg",
	}, "tool-type") {
		t.Fatalf("expected field targeted to a different custom asset type to be rejected")
	}
	if (DefinitionSet{serial, dose}).ValidateValues(map[string]any{"dose": "20mg"}) {
		t.Fatalf("expected targeted field to be rejected without an asset type")
	}
}

func TestDefinitionRequiresEnumOptionsOnlyForEnumFields(t *testing.T) {
	if _, ok := NewDefinition("definition-one", "tenant-one", "", ScopeTenant, "status", "Status", FieldTypeEnum, nil, ApplicabilityAllAssets, nil); ok {
		t.Fatalf("expected enum without options to be invalid")
	}
	option, ok := NewKey("new")
	if !ok {
		t.Fatalf("expected valid option")
	}
	if _, ok := NewDefinition("definition-one", "tenant-one", "", ScopeTenant, "serial", "Serial", FieldTypeText, []Key{option}, ApplicabilityAllAssets, nil); ok {
		t.Fatalf("expected non-enum with options to be invalid")
	}
}

func TestCustomAssetTypeArchiveLifecycle(t *testing.T) {
	assetType, ok := NewAssetType("type-one", "tenant-one", "inventory-one", ScopeInventory, "medicine", "Medicine", "")
	if !ok {
		t.Fatalf("expected valid custom asset type")
	}
	if assetType.LifecycleState != AssetTypeLifecycleActive || !assetType.IsActive() {
		t.Fatalf("expected active lifecycle state, got %+v", assetType)
	}

	archived, ok := assetType.Archive()
	if !ok {
		t.Fatalf("expected archive to succeed")
	}
	if archived.LifecycleState != AssetTypeLifecycleArchived || archived.IsActive() {
		t.Fatalf("expected archived lifecycle state, got %+v", archived)
	}
	if _, ok := archived.Archive(); ok {
		t.Fatalf("expected archiving an archived custom asset type to fail")
	}
}

func definition(t *testing.T, keyValue string, fieldType FieldType, rawOptions []string) Definition {
	t.Helper()

	id, ok := NewID("definition-" + keyValue)
	if !ok {
		t.Fatalf("expected valid id")
	}
	key, ok := NewKey(keyValue)
	if !ok {
		t.Fatalf("expected valid key %q", keyValue)
	}
	displayName, ok := NewDisplayName("Field " + keyValue)
	if !ok {
		t.Fatalf("expected valid display name")
	}
	options := make([]Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := NewKey(raw)
		if !ok {
			t.Fatalf("expected valid option %q", raw)
		}
		options = append(options, option)
	}
	definition, ok := NewDefinition(id, "tenant-one", "", ScopeTenant, key, displayName, fieldType, options, ApplicabilityAllAssets, nil)
	if !ok {
		t.Fatalf("expected valid definition")
	}
	return definition
}

func targetedDefinition(t *testing.T, keyValue string, fieldType FieldType, customAssetTypeID string) Definition {
	t.Helper()

	definition := definition(t, keyValue, fieldType, nil)
	targetID, ok := NewAssetTypeID(customAssetTypeID)
	if !ok {
		t.Fatalf("expected valid custom asset type id")
	}
	definition.Applicability = ApplicabilityCustomAssetTypes
	definition.CustomAssetTypeIDs = []AssetTypeID{targetID}
	return definition
}
