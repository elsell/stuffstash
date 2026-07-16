package asset

import "testing"

func TestKindIsPortable(t *testing.T) {
	tests := []struct {
		name     string
		kind     Kind
		portable bool
	}{
		{name: "item", kind: KindItem, portable: true},
		{name: "container", kind: KindContainer, portable: true},
		{name: "location", kind: KindLocation, portable: false},
		{name: "unknown", kind: Kind("unknown"), portable: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.kind.IsPortable(); got != test.portable {
				t.Fatalf("IsPortable() = %t, want %t", got, test.portable)
			}
		})
	}
}
