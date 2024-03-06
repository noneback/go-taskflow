package gotaskflow

import "reflect"

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
