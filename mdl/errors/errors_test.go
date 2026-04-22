// SPDX-License-Identifier: Apache-2.0

package mdlerrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrExit(t *testing.T) {
	err := ErrExit
	if !errors.Is(err, ErrExit) {
		t.Fatal("errors.Is(ErrExit, ErrExit) should be true")
	}
	wrapped := fmt.Errorf("wrapper: %w", ErrExit)
	if !errors.Is(wrapped, ErrExit) {
		t.Fatal("errors.Is(wrapped, ErrExit) should be true")
	}
}

func TestNotConnectedError(t *testing.T) {
	t.Run("read mode", func(t *testing.T) {
		err := NewNotConnected()
		if err.Error() != "not connected to a project" {
			t.Fatalf("unexpected message: %s", err.Error())
		}
		if err.WriteMode {
			t.Fatal("WriteMode should be false")
		}
		var target *NotConnectedError
		if !errors.As(err, &target) {
			t.Fatal("errors.As should match *NotConnectedError")
		}
	})

	t.Run("write mode", func(t *testing.T) {
		err := NewNotConnectedWrite()
		if err.Error() != "not connected to a project in write mode" {
			t.Fatalf("unexpected message: %s", err.Error())
		}
		if !err.WriteMode {
			t.Fatal("WriteMode should be true")
		}
		var target *NotConnectedError
		if !errors.As(err, &target) {
			t.Fatal("errors.As should match *NotConnectedError")
		}
	})

	t.Run("wrapped", func(t *testing.T) {
		inner := NewNotConnected()
		wrapped := fmt.Errorf("context: %w", inner)
		var target *NotConnectedError
		if !errors.As(wrapped, &target) {
			t.Fatal("errors.As should match through wrapping")
		}
	})
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFound("entity", "MyModule.MyEntity")
	if err.Error() != "entity not found: MyModule.MyEntity" {
		t.Fatalf("unexpected message: %s", err.Error())
	}
	if err.Kind != "entity" || err.Name != "MyModule.MyEntity" {
		t.Fatalf("unexpected fields: Kind=%s Name=%s", err.Kind, err.Name)
	}
	var target *NotFoundError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *NotFoundError")
	}

	custom := NewNotFoundMsg("microflow", "MyModule.DoSomething", "microflow MyModule.DoSomething does not exist")
	if custom.Error() != "microflow MyModule.DoSomething does not exist" {
		t.Fatalf("unexpected message: %s", custom.Error())
	}
}

func TestAlreadyExistsError(t *testing.T) {
	err := NewAlreadyExists("entity", "MyModule.MyEntity")
	if err.Error() != "entity already exists: MyModule.MyEntity" {
		t.Fatalf("unexpected message: %s", err.Error())
	}
	var target *AlreadyExistsError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *AlreadyExistsError")
	}

	custom := NewAlreadyExistsMsg("entity", "MyModule.MyEntity", "entity already exists: MyModule.MyEntity (use CREATE OR MODIFY to update)")
	if custom.Kind != "entity" {
		t.Fatalf("unexpected Kind: %s", custom.Kind)
	}
}

func TestUnsupportedError(t *testing.T) {
	err := NewUnsupported("unsupported attribute type: Binary")
	if err.Error() != "unsupported attribute type: Binary" {
		t.Fatalf("unexpected message: %s", err.Error())
	}
	var target *UnsupportedError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *UnsupportedError")
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidation("invalid entity name")
	if err.Error() != "invalid entity name" {
		t.Fatalf("unexpected message: %s", err.Error())
	}
	var target *ValidationError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *ValidationError")
	}

	errf := NewValidationf("invalid %s: %s", "entity name", "123Bad")
	if errf.Error() != "invalid entity name: 123Bad" {
		t.Fatalf("unexpected message: %s", errf.Error())
	}
}

func TestBackendError(t *testing.T) {
	cause := fmt.Errorf("disk full")
	err := NewBackend("write entity", cause)
	if err.Error() != "failed to write entity: disk full" {
		t.Fatalf("unexpected message: %s", err.Error())
	}
	var target *BackendError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *BackendError")
	}
	if target.Op != "write entity" {
		t.Fatalf("unexpected Op: %s", target.Op)
	}

	// Unwrap
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is should find the cause through Unwrap")
	}

	// Double-wrapped
	wrapped := fmt.Errorf("outer: %w", err)
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should match through double wrapping")
	}
	if !errors.Is(wrapped, cause) {
		t.Fatal("errors.Is should find cause through double wrapping")
	}

	// Nil cause
	nilErr := NewBackend("test op", nil)
	if nilErr.Error() != "failed to test op" {
		t.Fatalf("unexpected nil-cause message: %s", nilErr.Error())
	}
	if nilErr.Unwrap() != nil {
		t.Fatal("Unwrap should return nil when cause is nil")
	}
}
