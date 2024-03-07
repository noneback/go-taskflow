package gotaskflow

import (
	"reflect"
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

func unsafeToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func unsafeToBytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sliceHeader := reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
}
