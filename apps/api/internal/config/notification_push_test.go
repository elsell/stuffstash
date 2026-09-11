package config

import "testing"

func TestNotificationPushRequiresExplicitProviderConfiguration(t *testing.T) {
	cfg, err := LoadNotificationPush()
	if err != nil || cfg.Enabled() {
		t.Fatal("push enabled without credentials")
	}
	t.Setenv("STUFF_STASH_PUSH_APNS_ENABLED", "true")
	if _, err := LoadNotificationPush(); err == nil {
		t.Fatal("missing APNs settings accepted")
	}
	for k, v := range map[string]string{"SERVER_ID": "https://api.test", "APNS_KEY_ID": "ABCDEFGHIJ", "APNS_TEAM_ID": "KLMNOPQRST", "APNS_TOPIC": "org.stuffstash.mobile", "APNS_KEY_FILE": "/mounted/key.p8"} {
		t.Setenv("STUFF_STASH_PUSH_"+k, v)
	}
	cfg, err = LoadNotificationPush()
	if err != nil || !cfg.Enabled() || !cfg.APNSProduction {
		t.Fatalf("valid configuration rejected: %v", err)
	}
	t.Setenv("STUFF_STASH_PUSH_PAGE_TIMEOUT", "1m")
	if _, err := LoadNotificationPush(); err == nil {
		t.Fatal("page may outlive lease")
	}
}
