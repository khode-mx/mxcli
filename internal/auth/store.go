// SPDX-License-Identifier: Apache-2.0

package auth

// Store persists credentials across mxcli invocations.
//
// Implementations: fileStore (JSON on disk, mode 0600). An OS keychain
// implementation is planned but not required for v1.
type Store interface {
	// Get returns the credential for the given profile, or ErrNoCredential.
	Get(profile string) (*Credential, error)
	// Put stores a credential for the given profile, overwriting any existing one.
	Put(profile string, cred *Credential) error
	// Delete removes the credential for the given profile. Deleting a missing
	// profile is not an error.
	Delete(profile string) error
	// List returns all profile names with stored credentials.
	List() ([]string, error)
}
