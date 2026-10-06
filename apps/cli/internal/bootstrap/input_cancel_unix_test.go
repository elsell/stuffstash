//go:build !windows

package bootstrap

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInheritedCLIStdinCancellation(t *testing.T) {
	if os.Getenv("STUFF_STASH_TEST_INPUT_CHILD") == "1" {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		code := Run(ctx, []string{"tenants", "create", "--input", "-", "--no-input", "--server", "https://stash.example"}, os.Getenv, os.Stdout, os.Stderr)
		os.Exit(code)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestInheritedCLIStdinCancellation$")
	cmd.Env = append(os.Environ(), "STUFF_STASH_TEST_INPUT_CHILD=1", "STUFF_STASH_CLI_CONFIG_FILE="+filepath.Join(t.TempDir(), "config", "contexts.json"))
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var output bytes.Buffer
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	// Exceed the pipe capacity on the supported Unix hosts. Completion of
	// this write proves the child has started reading; keep the writer open.
	written := make(chan error, 1)
	go func() { _, err := io.WriteString(input, `{"name":"`+strings.Repeat("x", 512<<10)); written <- err }()
	select {
	case err := <-written:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CLI did not start reading stdin")
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 130 {
			t.Fatalf("exit=%v stderr=%s", err, &output)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("CLI remained blocked on inherited stdin after SIGINT")
	}
}
