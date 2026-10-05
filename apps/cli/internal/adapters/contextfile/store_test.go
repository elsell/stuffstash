package contextfile

import (
	"bufio"
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestContextFileRetainsConcurrentUpdates(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
	var wg sync.WaitGroup
	failures := make(chan error, 2)
	for _, name := range []string{"home", "work"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			failures <- store.Update(context.Background(), func(c *contexts.Config) error {
				c.Contexts = append(c.Contexts, contexts.Entry{Name: name, Server: "https://" + name + ".example"})
				c.Current = name
				return nil
			})
		}(name)
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := store.Load(context.Background())
	if err != nil || len(got.Contexts) != 2 || got.Version != contexts.Version {
		t.Fatalf("lost update: %+v %v", got, err)
	}
	file, err := os.Open(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if !secureFile(file) {
		t.Fatal("context file exposes scope")
	}

}
func TestContextFileRejectsSymlinkAndInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "other")
	if err := os.WriteFile(secret, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	store := Store{Path: filepath.Join(dir, "contexts.json")}
	if err := os.Symlink(secret, store.Path); err != nil {
		t.Skip("symlinks unavailable")
	}
	if _, err := store.Load(context.Background()); err == nil {
		t.Fatal("read symlink")
	}
	if err := store.Update(context.Background(), func(c *contexts.Config) error { return nil }); err == nil {
		t.Fatal("replaced symlink")
	}
	raw, _ := os.ReadFile(secret)
	if string(raw) != "untouched" {
		t.Fatal("changed unrelated file")
	}
	os.Remove(store.Path)
	if err := os.WriteFile(store.Path, []byte(`{"version":999,"contexts":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(context.Background()); err == nil {
		t.Fatal("accepted future version")
	}
}

func TestRejectedUpdatePreservesSavedContext(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
	if err := store.Update(context.Background(), func(c *contexts.Config) error {
		c.Current = "home"
		c.Contexts = []contexts.Entry{{Name: "home", Server: "https://home.example"}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*contexts.Config) error{
		func(c *contexts.Config) error { c.Contexts = append(c.Contexts, c.Contexts[0]); return nil },
		func(c *contexts.Config) error { c.Contexts = nil; return nil },
		func(c *contexts.Config) error { c.Contexts[0].Inventory = "orphan"; return nil },
	} {
		if err := store.Update(context.Background(), change); err == nil {
			t.Fatal("invalid update succeeded")
		}
		got, err := store.Load(context.Background())
		if err != nil || len(got.Contexts) != 1 || got.Current != "home" || got.Contexts[0].Inventory != "" {
			t.Fatalf("saved context corrupted: %+v %v", got, err)
		}
	}
}
func TestCanceledUpdateDoesNotCallChangeOrCreateDirectory(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "config", "contexts.json")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	if err := store.Update(ctx, func(c *contexts.Config) error { called = true; return nil }); err != context.Canceled {
		t.Fatalf("got %v", err)
	}
	if called {
		t.Fatal("called change after cancellation")
	}
	if _, err := os.Stat(filepath.Dir(store.Path)); !os.IsNotExist(err) {
		t.Fatalf("created directory after cancellation: %v", err)
	}
}

func TestContextLockProcessHelper(t *testing.T) {
	path := os.Getenv("STUFFSTASH_TEST_CONTEXT_LOCK")
	if path == "" {
		return
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	unlock, err := lock(context.Background(), root, filepath.Base(path)+".lock")
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err := os.Stdout.WriteString("locked\n"); err != nil {
		t.Fatal(err)
	}
	var b [1]byte
	_, _ = os.Stdin.Read(b[:])
}
func TestContextLockWaitAcrossProcessesHonorsCancellation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "config")
	if err := prepareDirectory(dir); err != nil {
		t.Fatal(err)
	}
	store := Store{Path: filepath.Join(dir, "contexts.json")}
	child := exec.Command(os.Args[0], "-test.run=^TestContextLockProcessHelper$")
	child.Env = append(os.Environ(), "STUFFSTASH_TEST_CONTEXT_LOCK="+store.Path)
	input, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		input.Close()
		if err := child.Wait(); err != nil {
			t.Errorf("lock holder: %v", err)
		}
	}()
	ready := make(chan string, 1)
	go func() { line, _ := bufio.NewReader(output).ReadString('\n'); ready <- line }()
	select {
	case line := <-ready:
		if line != "locked\n" {
			t.Fatalf("lock holder: %q", line)
		}
	case <-time.After(5 * time.Second):
		child.Process.Kill()
		t.Fatal("lock holder did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	called := false
	err = store.Update(ctx, func(c *contexts.Config) error { called = true; return nil })
	if !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatalf("lock wait: called=%v error=%v", called, err)
	}
	if _, err := os.Stat(store.Path); !os.IsNotExist(err) {
		t.Fatalf("published while another process held the lock: %v", err)
	}
}
