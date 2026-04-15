// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

const fileStoreVersion = 1

// fileFormat is the on-disk layout for ~/.mxcli/auth.json.
type fileFormat struct {
	Version  int                    `json:"version"`
	Profiles map[string]*Credential `json:"profiles"`
}

type fileStore struct {
	path string
}

// NewFileStore returns a Store backed by the given JSON file path.
// The file is created lazily on first Put.
func NewFileStore(path string) Store {
	return &fileStore{path: path}
}

// DefaultFileStore returns the file store at ~/.mxcli/auth.json.
func DefaultFileStore() (Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("auth: resolving home directory: %w", err)
	}
	return NewFileStore(filepath.Join(home, ".mxcli", "auth.json")), nil
}

func (s *fileStore) load() (*fileFormat, error) {
	info, err := os.Stat(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return &fileFormat{Version: fileStoreVersion, Profiles: map[string]*Credential{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth: stat %s: %w", s.path, err)
	}
	// Permission check — best effort on Windows, strict on Unix.
	if runtime.GOOS != "windows" {
		if mode := info.Mode().Perm(); mode&0o077 != 0 {
			return nil, &ErrPermissionsTooOpen{Path: s.path, Mode: mode}
		}
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("auth: read %s: %w", s.path, err)
	}
	if len(data) == 0 {
		return &fileFormat{Version: fileStoreVersion, Profiles: map[string]*Credential{}}, nil
	}
	var ff fileFormat
	if err := json.Unmarshal(data, &ff); err != nil {
		return nil, fmt.Errorf("auth: parse %s: %w", s.path, err)
	}
	if ff.Profiles == nil {
		ff.Profiles = map[string]*Credential{}
	}
	return &ff, nil
}

// saveAtomic writes the file via temp + rename. Leaves mode 0600 on create.
func (s *fileStore) saveAtomic(ff *fileFormat) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("auth: mkdir %s: %w", filepath.Dir(s.path), err)
	}
	data, err := json.MarshalIndent(ff, "", "  ")
	if err != nil {
		return fmt.Errorf("auth: encode: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".auth.json.*")
	if err != nil {
		return fmt.Errorf("auth: create temp: %w", err)
	}
	tmpPath := tmp.Name()
	// Clean up tmp on any error path below.
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("auth: write: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil && runtime.GOOS != "windows" {
		_ = tmp.Close()
		return fmt.Errorf("auth: chmod: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("auth: close temp: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("auth: rename %s -> %s: %w", tmpPath, s.path, err)
	}
	return nil
}

func (s *fileStore) Get(profile string) (*Credential, error) {
	ff, err := s.load()
	if err != nil {
		return nil, err
	}
	cred, ok := ff.Profiles[profile]
	if !ok {
		return nil, &ErrNoCredential{Profile: profile}
	}
	cred.Profile = profile
	return cred, nil
}

func (s *fileStore) Put(profile string, cred *Credential) error {
	if cred == nil {
		return fmt.Errorf("auth: nil credential")
	}
	if profile == "" {
		return fmt.Errorf("auth: empty profile name")
	}
	ff, err := s.load()
	if err != nil {
		// Permission errors are fatal; any other load error we can recover
		// from by starting fresh would risk overwriting a good file, so
		// surface it.
		return err
	}
	// Store a copy so we don't hold on to caller's pointer, and so Profile
	// (which is not serialized) is stable.
	stored := *cred
	stored.Profile = ""
	ff.Profiles[profile] = &stored
	return s.saveAtomic(ff)
}

func (s *fileStore) Delete(profile string) error {
	ff, err := s.load()
	if err != nil {
		if _, tooOpen := err.(*ErrPermissionsTooOpen); tooOpen {
			return err
		}
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		// Other errors are fatal.
		return err
	}
	if _, ok := ff.Profiles[profile]; !ok {
		return nil
	}
	delete(ff.Profiles, profile)
	return s.saveAtomic(ff)
}

func (s *fileStore) List() ([]string, error) {
	ff, err := s.load()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ff.Profiles))
	for name := range ff.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
