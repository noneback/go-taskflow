package utils

import (
	"sync/atomic"
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
