// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCredentialString_RedactsToken(t *testing.T) {
	c := &Credential{Scheme: SchemePAT, Token: "super-secret-token", CreatedAt: time.Now()}

	for _, verb := range []string{"%s", "%v", "%+v", "%#v"} {
		out := fmt.Sprintf(verb, c)
		if strings.Contains(out, "super-secret-token") {
			t.Errorf("format verb %q leaked token: %s", verb, out)
		}
		if !strings.Contains(out, "REDACTED") {
			t.Errorf("format verb %q missing REDACTED marker: %s", verb, out)
		}
	}
}

func TestCredentialString_NilSafe(t *testing.T) {
	var c *Credential
	if got := c.String(); got == "" {
		t.Errorf("nil credential String() should not be empty")
	}
}

func TestCredentialJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	in := &Credential{Scheme: SchemePAT, Token: "abc123", CreatedAt: now}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Profile must not be serialized.
	if strings.Contains(string(data), `"profile"`) {
		t.Errorf("profile leaked into JSON: %s", data)
	}
	var out Credential
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Scheme != in.Scheme || out.Token != in.Token || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("roundtrip mismatch: %+v vs %+v", in, out)
	}
}
