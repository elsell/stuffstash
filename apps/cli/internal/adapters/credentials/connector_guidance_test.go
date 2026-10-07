package credentials

import (
	"strings"
	"testing"
)

func TestConnectorStoreRecoveryRespectsPlatformSupport(t *testing.T) {
	for _, platform := range []string{"windows", "linux", "darwin"} {
		message := connectorStoreError("read", platform).Error()
		offersFile := strings.Contains(message, "STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE")
		if offersFile != (platform != "windows") {
			t.Fatalf("unsupported recovery on %s: %s", platform, message)
		}
		if !strings.Contains(message, "system credential store") {
			t.Fatalf("missing store guidance: %s", message)
		}
	}
}
