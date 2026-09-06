package config

import (
	"strings"
	"testing"
	"time"
)

func TestProfilingDisabledByDefault(t *testing.T) {
	t.Setenv("STUFF_STASH_PROFILING_ENABLED", "")
	cfg, err := LoadProfiling()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("profiling must require explicit enablement")
	}
}

func TestProfilingLoadsExplicitConnection(t *testing.T) {
	t.Setenv("STUFF_STASH_PROFILING_ENABLED", "true")
	t.Setenv("STUFF_STASH_PROFILING_ENDPOINT", "https://profiles.example.test")
	t.Setenv("STUFF_STASH_PROFILING_USERNAME", "test-stack")
	t.Setenv("STUFF_STASH_PROFILING_PASSWORD", "private-test-password")
	cfg, err := LoadProfiling()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Username != "test-stack" || cfg.Password != "private-test-password" || cfg.UploadInterval != 15*time.Second || cfg.RequestTimeout != 5*time.Second {
		t.Fatal("profile settings lost")
	}
}

func TestProfilingRejectsUnsafeConfigurationWithoutEchoingValues(t *testing.T) {
	for _, test := range []struct{ name, value string }{
		{"STUFF_STASH_PROFILING_ENABLED", "private-invalid"},
		{"STUFF_STASH_PROFILING_ENDPOINT", "https://user:private-secret@profiles.example.test"},
		{"STUFF_STASH_PROFILING_ENDPOINT", "file:///private-secret"},
		{"STUFF_STASH_PROFILING_UPLOAD_INTERVAL", "0s"},
		{"STUFF_STASH_PROFILING_REQUEST_TIMEOUT", "-1s"},
		{"STUFF_STASH_PROFILING_MUTEX_FRACTION", "-1"},
		{"STUFF_STASH_PROFILING_BLOCK_RATE", "-1"},
		{"PYROSCOPE_ADHOC_SERVER_ADDRESS", "https://private-secret.example.test"},
		{"PYROSCOPE_ADHOC_SERVER_ADDRESS", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("STUFF_STASH_PROFILING_ENABLED", "true")
			t.Setenv("STUFF_STASH_PROFILING_ENDPOINT", "https://profiles.example.test")
			t.Setenv(test.name, test.value)
			_, err := LoadProfiling()
			if err == nil {
				t.Fatal("unsafe profile settings accepted")
			}
			if strings.Contains(err.Error(), "private") {
				t.Fatal("configuration error exposed input")
			}
		})
	}
}
