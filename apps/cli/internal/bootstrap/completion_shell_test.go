package bootstrap

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The shell calls the actual local bootstrap through this test executable.
func TestCompletionHelperProcess(t *testing.T) {
	if os.Getenv("STUFFSTASH_COMPLETION_TEST_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Exit(Run(context.Background(), os.Args[i+1:], func(string) string { panic("completion read runtime environment") }, os.Stdout, os.Stderr))
		}
	}
	os.Exit(99)
}
func TestBashCompletionUsesLiteralArguments(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Native Bash completion is verified on the Linux host")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("Bash is not installed")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "completion.bash")
	var content, diagnostic bytes.Buffer
	if code := Run(context.Background(), []string{"completion", "bash"}, func(string) string { t.Fatal("generation read runtime environment"); return "" }, &content, &diagnostic); code != 0 {
		t.Fatalf("generate %d %s", code, &diagnostic)
	}
	if err := os.WriteFile(script, content.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(bash, "--noprofile", "--norc", "-n", script).CombinedOutput(); err != nil {
		t.Fatalf("Bash syntax: %v %s", err, output)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
	shim := "#!/bin/sh\nexec " + quote(executable) + " -test.run=^TestCompletionHelperProcess$ -- \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "stuffstash"), []byte(shim), 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "must-not-exist")
	for _, tc := range []struct {
		words        []string
		want, absent string
	}{
		{[]string{"stuffstash", "workflows", "revisions", "s"}, "show", "list"},
		{[]string{"stuffstash", "voice-provider", "show", "--"}, "--tenant", "--inventory"},
		{[]string{"stuffstash", "assets", "create", "--title", "$(touch " + marker + ")", "--ti"}, "--title", ""},
		{[]string{"stuffstash", "assets", "create", "--title", "workflows", "--pa"}, "--parent", ""},
		{[]string{"stuffstash", "--server", "https", ":", "//stash.example", "assets", "list", "--li"}, "--limit", ""},
		{[]string{"stuffstash", "--tenant", "=", "home", "workflows", "list", "--cu"}, "--cursor", ""},
		{[]string{"stuffstash", "assets", "show", "id", "--", "--"}, "", "--json"},
		{[]string{"stuffstash", "assets", "create", "--title", ""}, "", "--title"},
		{[]string{"stuffstash", "assets", "create", "--title", "=", "private"}, "", "--title"},
	} {
		code := `source "$1"
shift
COMP_WORDS=("$@")
COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
_stuffstash_complete
if ((${#COMPREPLY[@]})); then printf '%s\n' "${COMPREPLY[@]}"; fi
`
		args := append([]string{"--noprofile", "--norc", "-c", code, "completion-test", script}, tc.words...)
		cmd := exec.Command(bash, args...)
		cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "STUFFSTASH_COMPLETION_TEST_HELPER=1")
		output, err := cmd.CombinedOutput()
		if err != nil || tc.want != "" && !strings.Contains(string(output), tc.want) || tc.absent != "" && strings.Contains(string(output), tc.absent) || tc.want == "" && len(output) != 0 {
			t.Fatalf("%v: %v %s", tc.words, err, output)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("entered substitution was executed: %v", err)
	}
}
