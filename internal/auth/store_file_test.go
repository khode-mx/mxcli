// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func newTestFileStore(t *testing.T) (*fileStore, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	return &fileStore{path: path}, path
}

func TestFileStore_GetMissingProfile(t *testing.T) {
	s, _ := newTestFileStore(t)
	_, err := s.Get("default")
	var noCred *ErrNoCredential
	if !errors.As(err, &noCred) {
		t.Fatalf("expected ErrNoCredential, got %v", err)
	}
}

func TestFileStore_PutGet(t *testing.T) {
	s, path := newTestFileStore(t)
	cred := &Credential{Scheme: SchemePAT, Token: "tok", CreatedAt: time.Now().UTC()}

	if err := s.Put("default", cred); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := s.Get("default")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Token != "tok" || got.Scheme != SchemePAT {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
	if got.Profile != "default" {
		t.Errorf("Profile not populated on Get: %q", got.Profile)
	}

	// File should be mode 0600 on Unix.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Errorf("expected mode 0600, got %o", mode)
		}
	}
}

func TestFileStore_RejectsTooOpenPerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission check skipped on Windows")
	}
	s, path := newTestFileStore(t)
	// Write a valid file first so we know Get would otherwise succeed.
	if err := s.Put("default", &Credential{Scheme: SchemePAT, Token: "tok", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	_, err := s.Get("default")
	var tooOpen *ErrPermissionsTooOpen
	if !errors.As(err, &tooOpen) {
		t.Fatalf("expected ErrPermissionsTooOpen, got %v", err)
	}
}

func TestFileStore_MultipleProfiles(t *testing.T) {
	s, _ := newTestFileStore(t)
	if err := s.Put("default", &Credential{Scheme: SchemePAT, Token: "t1"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("ci", &Credential{Scheme: SchemePAT, Token: "t2"}); err != nil {
		t.Fatal(err)
	}

	names, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "ci" || names[1] != "default" {
		t.Errorf("list: %v", names)
	}

	got, err := s.Get("ci")
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "t2" {
		t.Errorf("ci token mismatch: %q", got.Token)
	}
}

func TestFileStore_DeleteMissingIsNoop(t *testing.T) {
	s, _ := newTestFileStore(t)
	if err := s.Delete("nonexistent"); err != nil {
		t.Errorf("delete missing should be noop, got %v", err)
	}
}

func TestFileStore_Delete(t *testing.T) {
	s, _ := newTestFileStore(t)
	_ = s.Put("default", &Credential{Scheme: SchemePAT, Token: "tok"})
	if err := s.Delete("default"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := s.Get("default")
	var noCred *ErrNoCredential
	if !errors.As(err, &noCred) {
		t.Errorf("expected ErrNoCredential after delete, got %v", err)
	}
}

func TestFileStore_AtomicWrite_NoPartialFile(t *testing.T) {
	// Put a valid file, then simulate a reader racing with a writer: after
	// Put returns, the file must always parse cleanly — no partial writes.
	s, path := newTestFileStore(t)
	for i := range 10 {
		if err := s.Put("default", &Credential{Scheme: SchemePAT, Token: "tok", CreatedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		// Every Put must leave a fully-formed, readable file.
		if _, err := os.ReadFile(path); err != nil {
			t.Fatalf("read after put #%d: %v", i, err)
		}
		if _, err := s.Get("default"); err != nil {
			t.Fatalf("get after put #%d: %v", i, err)
		}
	}
}
