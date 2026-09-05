package config

import "testing"

func TestLoadGoogleADCCredentialVersion(t *testing.T) {
	for _, value := range []string{"", " revision-42 "} {
		t.Setenv("STUFF_STASH_GOOGLE_ADC_CREDENTIAL_VERSION", value)
		expected := ""
		if value != "" {
			expected = "revision-42"
		}
		if got := Load().GoogleADCCredentialVersion; got != expected {
			t.Fatalf("ADC credential version = %q", got)
		}
	}
}
