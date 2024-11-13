package utils

import (
	"testing"
)

func TestUnsafeToString(t *testing.T) {
	original := "Hello, World!"
	b := []byte(original)

	s := UnsafeToString(b)

	if s != original {
		t.Errorf("Expected %q but got %q", original, s)
	}
}

func TestUnsafeToBytes(t *testing.T) {
	original := "Hello, World!"
	b := UnsafeToBytes(original)

	if string(b) != original {
		t.Errorf("Expected %q but got %q", original, string(b))
	}
}

func TestRC(t *testing.T) {
	rc := NewRC()

	if rc.Value() != 0 {
		t.Errorf("Expected count to be 0, got %d", rc.Value())
	}

	rc.Increase()
	if rc.Value() != 1 {
		t.Errorf("Expected count to be 1, got %d", rc.Value())
	}

	rc.Decrease()
	if rc.Value() != 0 {
		t.Errorf("Expected count to be 0, got %d", rc.Value())
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when decreasing below zero, but did not")
		}
	}()
	rc.Decrease()
}

func TestSet(t *testing.T) {
	rc := NewRC()
	rc.Set(5)

	if rc.Value() != 5 {
		t.Errorf("Expected count to be 5, got %d", rc.Value())
	}

	rc.Set(-1)

	if rc.Value() != -1 {
		t.Errorf("Expected count to be -1, got %d", rc.Value())
	}
}
