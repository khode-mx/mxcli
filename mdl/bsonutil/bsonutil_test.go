// SPDX-License-Identifier: Apache-2.0

package bsonutil

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/types"
)

func TestIDToBsonBinary_ValidUUID(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	bin := IDToBsonBinary(id)

	if bin.Subtype != 0x00 {
		t.Errorf("expected subtype 0x00, got 0x%02x", bin.Subtype)
	}
	if len(bin.Data) != 16 {
		t.Errorf("expected 16 bytes, got %d", len(bin.Data))
	}
}

func TestIDToBsonBinary_PanicsOnInvalidUUID(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on invalid UUID, got none")
		}
	}()
	IDToBsonBinary("not-a-uuid")
}

func TestIDToBsonBinary_PanicsOnEmptyString(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on empty string, got none")
		}
	}()
	IDToBsonBinary("")
}

func TestBsonBinaryToID_Roundtrip(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	bin := IDToBsonBinary(id)
	got := BsonBinaryToID(bin)
	if got != id {
		t.Errorf("roundtrip failed: got %q, want %q", got, id)
	}
}

func TestNewIDBsonBinary_ProducesValidUUID(t *testing.T) {
	bin := NewIDBsonBinary()
	if bin.Subtype != 0x00 {
		t.Errorf("expected subtype 0x00, got 0x%02x", bin.Subtype)
	}
	if len(bin.Data) != 16 {
		t.Errorf("expected 16 bytes, got %d", len(bin.Data))
	}

	// Convert back and validate UUID format
	id := BsonBinaryToID(bin)
	if !types.ValidateID(id) {
		t.Errorf("generated ID is not valid UUID format: %q", id)
	}
}

func TestNewIDBsonBinary_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := BsonBinaryToID(NewIDBsonBinary())
		if seen[id] {
			t.Fatalf("duplicate ID generated: %q", id)
		}
		seen[id] = true
	}
}

func TestIDToBsonBinaryErr_ValidUUID(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	bin, err := IDToBsonBinaryErr(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bin.Subtype != 0x00 {
		t.Errorf("expected subtype 0x00, got 0x%02x", bin.Subtype)
	}
	if len(bin.Data) != 16 {
		t.Errorf("expected 16 bytes, got %d", len(bin.Data))
	}
	// Roundtrip
	got := BsonBinaryToID(bin)
	if got != id {
		t.Errorf("roundtrip failed: got %q, want %q", got, id)
	}
}

func TestIDToBsonBinaryErr_InvalidUUID(t *testing.T) {
	_, err := IDToBsonBinaryErr("not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
}

func TestIDToBsonBinaryErr_EmptyString(t *testing.T) {
	_, err := IDToBsonBinaryErr("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}
