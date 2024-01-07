package gowk

import (
	"encoding/json"
	"reflect"
)

// 获取结构体
func CopyToStruct(old interface{}) interface{} {
	reflectType := reflect.TypeOf(old)
	sliceObj := reflect.New(reflectType.Elem())
	return sliceObj.Interface()
}

// 获取结构体切片
func CopyToStructSlice(old interface{}) interface{} {
	reflectType := reflect.TypeOf(old)
	sliceOfTtype := reflect.SliceOf(reflectType)
	sliceObj := reflect.New(sliceOfTtype)
	sliceObj.Elem().Set(reflect.Zero(sliceOfTtype))
	return sliceObj.Interface()
}

func CopyByJson[O any, N any](o *O) *N {
	bytes, _ := json.Marshal(o)
	var n N
	json.Unmarshal(bytes, &n)
	return &n
}
