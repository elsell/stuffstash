package config

import "testing"

func TestExportLimitsRejectInvalidConfiguration(t *testing.T) {
	for _, name := range []string{"STUFF_STASH_EXPORT_MAX_RECORDS", "STUFF_STASH_EXPORT_MAX_BYTES"} {
		for _, value := range []string{"0", "-1", "invalid"} {
			t.Run(name+value, func(t *testing.T) {
				t.Setenv(name, value)
				if _, _, err := Load().Exports.Limits(); err == nil {
					t.Fatal("invalid export bound accepted")
				}
			})
		}
	}
	t.Setenv("STUFF_STASH_EXPORT_MAX_RECORDS", "25")
	t.Setenv("STUFF_STASH_EXPORT_MAX_BYTES", "1024")
	records, bytes, err := Load().Exports.Limits()
	if err != nil || records != 25 || bytes != 1024 {
		t.Fatalf("limits: %d %d %v", records, bytes, err)
	}
}
