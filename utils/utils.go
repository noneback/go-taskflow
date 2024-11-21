package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func UnsafeToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func UnsafeToBytes(s string) []byte {
	if len(s) == 0 {
		return []byte{}
	}

	ptr := unsafe.StringData(s)
	return unsafe.Slice(ptr, len(s))
}

// Reference Counter
type RC struct {
	cnt   int
	mutex *sync.Mutex
}

func NewRC() *RC {
	return &RC{
		0, &sync.Mutex{},
	}
}

func (c *RC) Increase() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cnt++
}

func (c *RC) Decrease() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cnt < 1 {
		panic("RC cannot be negetive")
	}
	c.cnt--
}

func (c *RC) Value() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.cnt
}

func (c *RC) Set(val int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cnt = val
}

// NormalizeDuration normalize duration
func NormalizeDuration(d time.Duration) string {
	ns := d.Nanoseconds()
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000
	microseconds := int(d.Microseconds()) % 1000
	nanoseconds := ns % int64(time.Microsecond)

	var parts []byte

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours)...)
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes)...)
	}
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds)...)
	}
	if milliseconds > 0 {
		parts = append(parts, fmt.Sprintf("%dms", milliseconds)...)
	}
	if microseconds > 0 {
		parts = append(parts, fmt.Sprintf("%dµs", microseconds)...)
	}
	if nanoseconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dns", nanoseconds)...)
	}

	return UnsafeToString(parts)
}

func AssertPanics(t *testing.T, name string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: didn't panic as expected", name)
		}
	}()

	f()
}
