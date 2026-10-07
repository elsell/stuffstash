package bootstrap

import (
	"bytes"
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPrintConfigurationErrorsIdentifyControlsBeforeExternalActions(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		action string
		wants  []string
	}{
		{"backoff", map[string]string{"STUFF_STASH_CLI_PRINT_BACKOFF_MIN": "10s", "STUFF_STASH_CLI_PRINT_BACKOFF_MAX": "1s"}, "run", []string{"STUFF_STASH_CLI_PRINT_BACKOFF_MIN", "STUFF_STASH_CLI_PRINT_BACKOFF_MAX"}},
		{"bytes", map[string]string{"STUFF_STASH_CLI_PRINT_MAX_ARTIFACT_BYTES": "67108865"}, "run", []string{"STUFF_STASH_CLI_PRINT_MAX_ARTIFACT_BYTES", "67108864"}},
		{"pairing", map[string]string{"STUFF_STASH_CLI_PAIRING_POLL_INTERVAL": "0s"}, "rotate", []string{"STUFF_STASH_CLI_PAIRING_POLL_INTERVAL", "1s", "1m"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.action == "run" && runtime.GOOS != "linux" {
				t.Skip("worker configuration requires Linux")
			}
			config := filepath.Join(t.TempDir(), "contexts.json")
			getenv := func(k string) string {
				if k == "STUFF_STASH_CLI_CONFIG_FILE" {
					return config
				}
				return test.values[k]
			}
			var out, diagnostic bytes.Buffer
			code := Run(context.Background(), []string{"connectors", "print", test.action, "--connector", "test", "--server", "https://stash.invalid", "--no-input", "--json"}, getenv, &out, &diagnostic)
			if code != 2 || !strings.Contains(diagnostic.String(), `"category":"configuration"`) {
				t.Fatalf("expected local configuration rejection: %d %s", code, &diagnostic)
			}
			for _, want := range test.wants {
				if !strings.Contains(diagnostic.String(), want) {
					t.Errorf("missing control %s: %s", want, &diagnostic)
				}
			}
		})
	}
}
