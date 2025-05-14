package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCPUProfile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cpu_profile.*.prof")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	profiler := NewPprofUtils(CPU, tmpFile.Name())
	defer profiler.StopProfile()

	profiler.StartProfile()

	doCPUWork()
}

func TestHeapProfile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "heap_profile.*.prof")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	profiler := NewPprofUtils(HEAP, tmpFile.Name())
	defer profiler.StopProfile()
	profiler.StartProfile()

	doMemoryWork()
}

func TestInvalidProfileType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid profile type, but got none")
		}
	}()

	p := NewPprofUtils(ProfileType(99), "invalid.prof")
	p.StartProfile()
	defer p.StopProfile()
}

func TestFileCreateError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when file creation fails, but got none")
		}
	}()

	_ = NewPprofUtils(CPU, "/nonexistent/cpu.prof")
}

func doCPUWork() {
	for i := 0; i < 1e6; i++ {
		_ = i * i
	}
}

func doMemoryWork() {
	data := make([]byte, 10<<20)
	_ = data
}
