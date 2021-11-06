package gowk

import "reflect"

// 获取结构体
func CopyToStruct(old interface{}) interface{} {
	reflectType := reflect.TypeOf(old)
	sliceObj := reflect.New(reflectType.Elem())
	return sliceObj.Interface()
}

//获取结构体切片
func CopyToStructSlice(old interface{}) interface{} {
	reflectType := reflect.TypeOf(old)
	sliceOfTtype := reflect.SliceOf(reflectType)
	sliceObj := reflect.New(sliceOfTtype)
	sliceObj.Elem().Set(reflect.Zero(sliceOfTtype))
	return sliceObj.Interface()
}
