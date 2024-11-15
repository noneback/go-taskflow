package utils

import (
	"fmt"
	"testing"
	"time"
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

func TestPanic(t *testing.T) {
	f := func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in causePanic:", r)
			}
			fmt.Println("1")
		}()

		fmt.Println("result")
	}
	f()
}

func TestNormalizeDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{time.Duration(0), "0ns"},
		{time.Second, "1s"},
		{time.Duration(2 * time.Second), "2s"},
		{time.Minute, "1m"},
		{time.Duration(61 * time.Second), "1m1s"},
		{time.Hour, "1h"},
		{time.Duration(3601 * time.Second), "1h1s"},
		{time.Duration(1*time.Hour + 30*time.Minute + 15*time.Second), "1h30m15s"},
		{time.Duration(1*time.Minute + 500*time.Millisecond), "1m500ms"},
		{time.Duration(500 * time.Microsecond), "500µs"},
		{time.Duration(1 * time.Nanosecond), "1ns"},
	}

	for _, test := range tests {
		t.Run(test.input.String(), func(t *testing.T) {
			result := NormalizeDuration(test.input)
			if result != test.expected {
				t.Errorf("NormalizeDuration(%v) = %v; want %v", test.input, result, test.expected)
			}
		})
	}
}
