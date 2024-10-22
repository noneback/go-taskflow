package utils

import (
	"fmt"
	"sync/atomic"
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
	cnt atomic.Int32
}

func (c *RC) Increase() {
	c.cnt.Add(1)
}

func (c *RC) Decrease() {
	if c.cnt.Load() < 1 {
		panic("RC cannot be negetive")
	}
	c.cnt.Add(-1)
}

func (c *RC) Value() int {
	return int(c.cnt.Load())
}

func (c *RC) Set(val int) {
	c.cnt.Store(int32(val))
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
