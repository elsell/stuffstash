package config

import "testing"

func TestArchiveConfigurationBoundsAndDurableMode(t *testing.T) {
	cfg, err := LoadArchives("postgres")
	if err != nil || !cfg.Enabled {
		t.Fatal("persistent archive defaults")
	}
	cfg, err = LoadArchives("memory")
	if err != nil || cfg.Enabled {
		t.Fatal("memory advertised durable archives")
	}
	for _, entry := range []struct{ name, value string }{{"MAX_BYTES", "0"}, {"ENTRY_BYTES", "9999999999999"}, {"HEARTBEAT", "2m"}, {"CONCURRENCY", "9"}, {"RETENTION", "1s"}, {"JOB_TIMEOUT", "0s"}, {"MAX_RECORDS", "-1"}} {
		t.Run(entry.name, func(t *testing.T) {
			t.Setenv("STUFF_STASH_ARCHIVE_"+entry.name, entry.value)
			if _, err := LoadArchives("postgres"); err == nil {
				t.Fatal("invalid archive configuration accepted")
			}
		})
	}
	t.Setenv("STUFF_STASH_ARCHIVE_ENABLED", "true")
	if _, err = LoadArchives("memory"); err == nil {
		t.Fatal("durable worker on ephemeral store")
	}
}
