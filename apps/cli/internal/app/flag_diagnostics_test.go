package app

import (
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
	"testing"
)

func TestFlagConversionErrorsIdentifyOptionWithoutValue(t *testing.T) {
	for _, tc := range []struct{ flag, kind string }{{"limit", "integer"}, {"copies", "integer"}, {"template-version", "non-negative integer"}, {"width-mm", "number"}, {"json", "true or false"}} {
		t.Run(tc.flag, func(t *testing.T) {
			secret := "private-token-\x1b[31m"
			_, err := Parse([]string{"assets", "list", "--" + tc.flag + "=" + secret}, func(string) string { return "" })
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "usage" {
				t.Fatalf("wrong failure: %v", err)
			}
			for _, part := range []string{"--" + tc.flag, tc.kind, "Supply"} {
				if !strings.Contains(failure.Message, part) {
					t.Errorf("missing %s: %s", part, failure.Message)
				}
			}
			if strings.Contains(failure.Message, "private-token") || strings.Contains(failure.Message, "\x1b") {
				t.Fatal("value exposed")
			}
		})
	}
}
func TestFlagRangesNameIndependentBounds(t *testing.T) {
	for _, tc := range []struct {
		args        []string
		want, other string
	}{
		{[]string{"labels", "print", "asset", "--copies", "0"}, "--copies", "template-version"},
		{[]string{"labels", "print", "asset", "--template-version", "4294967296"}, "4294967295", "copies"},
	} {
		_, err := Parse(tc.args, func(string) string { return "" })
		if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), tc.other) {
			t.Fatalf("range: %v", err)
		}
	}
}
func TestDiagnosticParsingPreservesValuesAndExplicitFlags(t *testing.T) {
	o, err := Parse([]string{"labels", "render", "asset", "--output", "label.png", "--width-mm", "-1.25", "--height-mm=2.5", "--copies", "2", "--copies=3", "--show-reference=true", "--show-reference=false", "--json=false", "--tenant", "home"}, func(k string) string {
		if k == "STUFF_STASH_CLI_TENANT" {
			return "old"
		}
		return ""
	})
	if err != nil || o.WidthMM != -1.25 || o.HeightMM != 2.5 || o.Copies != 3 || !o.ShowReferenceSet || o.ShowReference || o.JSON || o.Selection.Tenant != "home" {
		t.Fatalf("semantics changed: %+v %v", o, err)
	}
	// Visited explicit false options still take part in command-specific validation.
	if _, err := Parse([]string{"workflows", "list", "--yes=false"}, func(string) string { return "" }); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("explicit false forgotten: %v", err)
	}
}
