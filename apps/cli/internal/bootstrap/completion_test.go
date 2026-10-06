package bootstrap

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCompletionGenerationIsLocal(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		var out, diag bytes.Buffer
		code := Run(context.Background(), []string{"completion", shell, "--json", "--no-input"}, func(string) string { t.Fatal("completion read runtime environment"); return "" }, &out, &diag)
		if code != 0 || diag.Len() != 0 || !strings.Contains(out.String(), "command stuffstash __complete "+shell+" --") {
			t.Fatalf("generation %s: %d %s %s", shell, code, &out, &diag)
		}
		for _, unsafe := range []string{"eval ", "bash -c", "sh -c", "curl ", "wget "} {
			if strings.Contains(out.String(), unsafe) {
				t.Errorf("unsafe script: %s", unsafe)
			}
		}
	}
	var out, diag bytes.Buffer
	if code := Run(context.Background(), []string{"completion", "unknown"}, func(string) string { t.Fatal("unknown shell read config"); return "" }, &out, &diag); code != 2 {
		t.Fatalf("unknown shell %d", code)
	}
}
func TestCompletionQueryIsStaticAndContextual(t *testing.T) {
	for _, tc := range []struct {
		current      string
		prior        []string
		want, absent string
	}{
		{"work", nil, "workflows", "assets"},
		{"", []string{"workflows", "revisions"}, "show", "assets"},
		{"--", []string{"voice-provider", "show"}, "--tenant", "--inventory"},
		{"--", []string{"import-jobs", "list"}, "--inventory", "--cursor"},
		{"--ti", []string{"--tenant", "workflows", "assets", "create"}, "--title", ""},
		{"--", []string{"assets", "show", "private-id"}, "--json", "private-id"},
		{"", []string{"assets", "create", "--title"}, "", "--title"},
		{"", []string{"assets", "create", "--title", "="}, "", "--title"},
		{"--", []string{"assets", "show", "id", "--"}, "", "--json"},
		{"--title=private", []string{"assets", "create"}, "", "--title"},
		{"--", []string{"assets", "invented"}, "", "--json"},
	} {
		args := []string{"__complete", "bash", "--", tc.current}
		args = append(args, tc.prior...)
		var out, diag bytes.Buffer
		code := Run(context.Background(), args, func(string) string { t.Fatal("query read runtime state"); return "" }, &out, &diag)
		if code != 0 || diag.Len() != 0 || tc.want != "" && !strings.Contains(out.String(), tc.want) || tc.absent != "" && strings.Contains(out.String(), tc.absent) || tc.want == "" && out.Len() != 0 {
			t.Fatalf("%v: %d %s %s", args, code, &out, &diag)
		}
	}
}
