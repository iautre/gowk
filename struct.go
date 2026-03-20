package gowk

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// CopyToStruct 根据传入的指针值创建一个同类型的新零值指针。
// 传入非指针类型时返回 error 而非 panic。
func CopyToStruct(old interface{}) (interface{}, error) {
	if old == nil {
		return nil, fmt.Errorf("CopyToStruct: 参数不能为 nil")
	}
	t := reflect.TypeOf(old)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("CopyToStruct: 期望指针类型，得到 %s", t.Kind())
	}
	return reflect.New(t.Elem()).Interface(), nil
}

// CopyToStructSlice 根据传入的指针值创建一个同类型的空切片指针。
func CopyToStructSlice(old interface{}) interface{} {
	reflectType := reflect.TypeOf(old)
	sliceOfTtype := reflect.SliceOf(reflectType)
	sliceObj := reflect.New(sliceOfTtype)
	sliceObj.Elem().Set(reflect.Zero(sliceOfTtype))
	return sliceObj.Interface()
}

// CopyByJson 通过 JSON 序列化/反序列化将 O 类型转换为 N 类型。
// 任一步骤失败时返回零值和 error。
func CopyByJson[O any, N any](o O) (N, error) {
	var n N
	b, err := json.Marshal(o)
	if err != nil {
		return n, fmt.Errorf("CopyByJson marshal: %w", err)
	}
	if err := json.Unmarshal(b, &n); err != nil {
		return n, fmt.Errorf("CopyByJson unmarshal: %w", err)
	}
	return n, nil
}
