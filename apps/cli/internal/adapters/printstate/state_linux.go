//go:build linux

package printstate

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/sys/unix"
)

const journalVersion = 1
const maximumJournalBytes = 64 << 10

type envelope struct {
	Version  int                     `json:"version"`
	Device   string                  `json:"device"`
	Record   *printing.JournalRecord `json:"record"`
	Checksum string                  `json:"checksum"`
}
type locked struct {
	mu           sync.Mutex
	root         *os.Root
	lock         *os.File
	device, name string
	closed       bool
}

func private(info os.FileInfo, directory bool) error {
	metadata, ok := info.Sys().(*syscall.Stat_t)
	if !ok || metadata.Uid != uint32(os.Getuid()) || info.Mode().Perm()&0077 != 0 || directory != info.IsDir() || !directory && !info.Mode().IsRegular() {
		return errors.New("print state must be owned by the current user and owner-only")
	}
	return nil
}
func (s Store) Acquire(ctx context.Context, device string) (ports.LockedPrintState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Directory == "" || device == "" {
		return nil, errors.New("print state directory and physical device identity are required")
	}
	if err := os.MkdirAll(s.Directory, 0700); err != nil {
		return nil, err
	}
	initial, err := os.Lstat(s.Directory)
	if err != nil {
		return nil, err
	}
	if err = private(initial, true); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(s.Directory)
	if err != nil {
		return nil, err
	}
	closeRoot := true
	defer func() {
		if closeRoot {
			root.Close()
		}
	}()
	current, err := root.Stat(".")
	if err != nil {
		return nil, err
	}
	if !os.SameFile(initial, current) {
		return nil, errors.New("print state directory changed while opening")
	}
	identity := sha256.Sum256([]byte(device))
	name := hex.EncodeToString(identity[:])
	// Root confines paths; O_NOFOLLOW refuses a substituted lock symlink.
	file, err := root.OpenFile(name+".lock", os.O_RDWR|os.O_CREATE|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	closeFile := true
	defer func() {
		if closeFile {
			file.Close()
		}
	}()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if err = private(info, false); err != nil {
		return nil, err
	}
	if err = unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, ports.ErrDeviceInUse
		}
		return nil, err
	}
	closeRoot = false
	closeFile = false
	return &locked{root: root, lock: file, device: device, name: name + ".json"}, nil
}
func checksum(value envelope) (string, error) {
	value.Checksum = ""
	body, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}
func (s *locked) Load(ctx context.Context) (*printing.JournalRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.closed {
		return nil, os.ErrClosed
	}
	file, err := s.root.OpenFile(s.name, os.O_RDONLY|unix.O_NOFOLLOW, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ports.ErrJournalMissing
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if err = private(info, false); err != nil {
		return nil, err
	}
	if info.Size() > maximumJournalBytes {
		return nil, ports.ErrJournalCorrupt
	}
	decoder := json.NewDecoder(io.LimitReader(file, maximumJournalBytes+1))
	decoder.DisallowUnknownFields()
	var value envelope
	if err = decoder.Decode(&value); err != nil {
		return nil, ports.ErrJournalCorrupt
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, ports.ErrJournalCorrupt
	}
	expected, err := checksum(value)
	if err != nil || value.Version != journalVersion || value.Device != s.device || value.Checksum != expected {
		return nil, ports.ErrJournalCorrupt
	}
	if value.Record != nil && value.Record.Validate() != nil {
		return nil, ports.ErrJournalCorrupt
	}
	return value.Record, nil
}
func (s *locked) Save(ctx context.Context, record *printing.JournalRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.closed {
		return os.ErrClosed
	}
	if record != nil {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	value := envelope{Version: journalVersion, Device: s.device, Record: record}
	sum, err := checksum(value)
	if err != nil {
		return err
	}
	value.Checksum = sum
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(body) > maximumJournalBytes {
		return errors.New("print recovery evidence exceeds journal limit")
	}
	var entropy [16]byte
	if _, err = rand.Read(entropy[:]); err != nil {
		return err
	}
	temporary := fmt.Sprintf(".%s-%x", s.name, entropy)
	file, err := s.root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer s.root.Remove(temporary)
	if _, err = file.Write(body); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = s.root.Rename(temporary, s.name); err != nil {
		return err
	}
	directory, err := s.root.Open(".")
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
func (s *locked) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return errors.Join(s.lock.Close(), s.root.Close())
}
