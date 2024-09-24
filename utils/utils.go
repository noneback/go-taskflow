package utils

import (
	"reflect"
	"sync/atomic"
	"unsafe"
)

func Convert[T any](in interface{}) (T, bool) {
	var tmp T
	inType := reflect.TypeOf(in)
	targetType := reflect.TypeOf(tmp)
	if inType.ConvertibleTo(targetType) {
		val := reflect.ValueOf(in)
		converted := val.Convert(targetType)
		return converted.Interface().(T), true
	}
	return tmp, false
}

func UnsafeToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func UnsafeToBytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sliceHeader := reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
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
