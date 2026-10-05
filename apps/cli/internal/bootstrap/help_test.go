package bootstrap

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCommandHelpIsLocalAndRelevant(t *testing.T) {
	for _, tc := range []struct {
		args         []string
		want, absent []string
	}{
		{[]string{"assets", "show", "private-resource-id", "--help", "--json", "--no-input"}, []string{"stuffstash assets show ASSET_ID", "Scope:", "--inventory", "Output:", "Example:"}, []string{"private-resource-id", "--journal-dir", "--title"}},
		{[]string{"import-jobs", "list", "--help"}, []string{"List import jobs.", "--inventory"}, []string{"--limit", "--cursor"}},
		{[]string{"assets", "create", "--input", "/does/not/exist", "--help"}, []string{"--kind", "--title", "--input", "Input:", "Confirmation:", "Example:"}, []string{"--journal-dir", "--unread-only"}},
		{[]string{"print-jobs", "resolve", "--help"}, []string{"revision", "acknowledgeUncertainty", "--yes", "--input"}, []string{"--title"}},
		{[]string{"context", "delete", "--help"}, []string{"stuffstash context delete NAME", "Local", "Confirmation:"}, []string{"--tenant", "--inventory"}},
		{[]string{"voice-provider", "show", "--help"}, []string{"--tenant", "household", "Output:"}, []string{"--inventory", "--limit", "--input"}},
		{[]string{"workflows", "revisions", "show", "--help"}, []string{"WORKFLOW_ID REVISION_ID", "--tenant"}, []string{"--limit", "--inventory"}},
		{[]string{"connectors", "print", "run", "--help"}, []string{"--connector", "--journal-dir", "connector", "long-running"}, []string{"--title"}},
		{[]string{"printers", "discover", "--help"}, []string{"Local", "Example:"}, []string{"--tenant", "--inventory", "--server"}},
	} {
		t.Run(strings.Join(tc.args[:2], "_"), func(t *testing.T) {
			var out, diag bytes.Buffer
			getenv := func(string) string { t.Fatal("help read runtime environment"); return "" }
			code := Run(context.Background(), tc.args, getenv, &out, &diag)
			if code != 0 || diag.Len() != 0 {
				t.Fatalf("help %d %s %s", code, &out, &diag)
			}
			for _, s := range tc.want {
				if !strings.Contains(out.String(), s) {
					t.Errorf("missing %q in %s", s, &out)
				}
			}
			for _, s := range tc.absent {
				if strings.Contains(out.String(), s) {
					t.Errorf("irrelevant %q in %s", s, &out)
				}
			}
		})
	}
}

func TestHelpNavigationAndUnknownActions(t *testing.T) {
	for _, tc := range []struct {
		args         []string
		want, absent string
		code         int
	}{
		{[]string{"--help"}, "assets", "assets return-details", 0},
		{[]string{"workflows", "--help"}, "revisions", "labels render", 0},
		{[]string{"workflows", "revisions", "--help"}, "show", "notification-devices", 0},
		{[]string{"assets", "invented-action", "--help"}, "", "", 2},
	} {
		var out, diag bytes.Buffer
		code := Run(context.Background(), tc.args, func(string) string { t.Fatal("help read configuration"); return "" }, &out, &diag)
		if code != tc.code || tc.want != "" && !strings.Contains(out.String(), tc.want) || tc.absent != "" && strings.Contains(out.String(), tc.absent) {
			t.Fatalf("%v: %d %s %s", tc.args, code, &out, &diag)
		}
	}
}
