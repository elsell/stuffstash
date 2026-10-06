package bootstrap

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFlagDiagnosticsKeepUsageExitAndHideValues(t *testing.T) {
	for _, args := range [][]string{{"assets", "list", "--limit=private-token"}, {"--json", "assets", "list", "--limit=private-token"}, {"assets", "list", "--limit=private-token", "--help"}, {"--json=private-token", "--help"}} {
		var out, diag bytes.Buffer
		code := Run(context.Background(), args, func(string) string { return "" }, &out, &diag)
		if code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private-token") || !strings.Contains(diag.String(), "Supply") {
			t.Fatalf("unsafe/nonactionable %d %s %s", code, &out, &diag)
		}
	}
}
