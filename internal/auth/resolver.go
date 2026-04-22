// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"os"
	"strings"
	"time"
)

// Env var names checked by the resolver.
const (
	EnvPAT     = "MENDIX_PAT"
	EnvProfile = "MXCLI_PROFILE"
)

// envProfile returns the profile name env vars should populate, or
// ProfileDefault if MXCLI_PROFILE is unset.
func envProfile() string {
	if p := strings.TrimSpace(os.Getenv(EnvProfile)); p != "" {
		return p
	}
	return ProfileDefault
}

// credentialFromEnv returns a Credential synthesized from env vars, or nil
// if no env-var credential is present.
func credentialFromEnv(profile string) *Credential {
	if envProfile() != profile {
		return nil
	}
	if pat := strings.TrimSpace(os.Getenv(EnvPAT)); pat != "" {
		return &Credential{
			Profile:   profile,
			Scheme:    SchemePAT,
			Token:     pat,
			CreatedAt: time.Now().UTC(),
		}
	}
	return nil
}

// Resolve returns the credential for the given profile. Env vars take
// precedence over the store.
//
// Returns *ErrNoCredential when no credential is available.
func Resolve(ctx context.Context, profile string) (*Credential, error) {
	if profile == "" {
		profile = ProfileDefault
	}
	if cred := credentialFromEnv(profile); cred != nil {
		return cred, nil
	}
	store, err := DefaultFileStore()
	if err != nil {
		return nil, err
	}
	return store.Get(profile)
}
