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

func TestDefinitionRequiresEnumOptionsOnlyForEnumFields(t *testing.T) {
	if _, ok := NewDefinition("definition-one", "tenant-one", "", ScopeTenant, "status", "Status", FieldTypeEnum, nil); ok {
		t.Fatalf("expected enum without options to be invalid")
	}
	option, ok := NewKey("new")
	if !ok {
		t.Fatalf("expected valid option")
	}
	if _, ok := NewDefinition("definition-one", "tenant-one", "", ScopeTenant, "serial", "Serial", FieldTypeText, []Key{option}); ok {
		t.Fatalf("expected non-enum with options to be invalid")
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
	definition, ok := NewDefinition(id, "tenant-one", "", ScopeTenant, key, displayName, fieldType, options)
	if !ok {
		t.Fatalf("expected valid definition")
	}
	return definition
}
