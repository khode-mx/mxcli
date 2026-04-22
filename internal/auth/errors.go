// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"os"
)

// ErrUnauthenticated is returned when a stored credential is rejected by the
// platform (HTTP 401). Callers can check with errors.As to show a hint
// pointing at "mxcli auth login".
type ErrUnauthenticated struct {
	Profile string
}

func (e *ErrUnauthenticated) Error() string {
	if e.Profile == "" {
		return "auth: credential was rejected (401). Run: mxcli auth login"
	}
	return fmt.Sprintf("auth: credential for profile %q was rejected (401). Run: mxcli auth login --profile %s", e.Profile, e.Profile)
}

// ErrSchemeMismatch is returned when a request targets a host whose required
// auth scheme does not match the resolved credential's scheme.
type ErrSchemeMismatch struct {
	Host string
	Need Scheme
	Have Scheme
}

func (e *ErrSchemeMismatch) Error() string {
	return fmt.Sprintf("auth: host %s requires scheme %q but credential has scheme %q", e.Host, e.Need, e.Have)
}

// ErrNoCredential is returned when no credential is available for a profile
// (neither in env vars nor in the store).
type ErrNoCredential struct {
	Profile string
}

func (e *ErrNoCredential) Error() string {
	return fmt.Sprintf("auth: no credential for profile %q. Run: mxcli auth login --profile %s", e.Profile, e.Profile)
}

// ErrPermissionsTooOpen is returned when the credential file has mode bits
// broader than 0600. Prevents accidental credential disclosure on shared
// systems.
type ErrPermissionsTooOpen struct {
	Path string
	Mode os.FileMode
}

func (e *ErrPermissionsTooOpen) Error() string {
	return fmt.Sprintf("auth: credential file %s has permissions %o (must be 0600). Run: chmod 0600 %s", e.Path, e.Mode.Perm(), e.Path)
}

// ErrUnknownHost is returned when a request targets a host that is not a
// known Mendix platform endpoint (not in the host→scheme map).
type ErrUnknownHost struct {
	Host string
}

func (e *ErrUnknownHost) Error() string {
	return fmt.Sprintf("auth: unknown Mendix host %q — refusing to send credentials", e.Host)
}
