package bootstrap

import (
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDisabledPushDoesNotReadCredentials(t *testing.T) {
	sender, err := buildNotificationPush(config.NotificationPushConfig{APNSKeyFile: "/missing/private-key"}, ports.SystemClock{})
	if err != nil || sender != nil {
		t.Fatal("disabled provider accessed credentials")
	}
}
func TestPushCredentialFailuresDoNotExposeFileOrContents(t *testing.T) {
	for _, contents := range []string{"", "secret-invalid-key", strings.Repeat("x", 65537)} {
		path := filepath.Join(t.TempDir(), "private-key")
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := buildNotificationPush(config.NotificationPushConfig{APNSEnabled: true, APNSKeyFile: path}, ports.SystemClock{})
		if err != errPushCredentials {
			t.Fatalf("expected fixed credential failure, got %v", err)
		}
	}
	_, err := readPushCredential(filepath.Join(t.TempDir(), "missing"))
	if err != errPushCredentials {
		t.Fatal("missing credential did not return fixed error")
	}
}
