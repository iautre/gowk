package gowk

import (
	"net/http"
	"reflect"
)

// Struct2Map 将结构体转换为 map，优先使用 json tag，无 tag 时使用字段名。
// 仅处理导出字段；传入非结构体时返回空 map。
func Struct2Map(obj any) (map[string]any, error) {
	data := make(map[string]any)
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	if objT == nil || objT.Kind() != reflect.Struct {
		return data, nil
	}
	for i := 0; i < objT.NumField(); i++ {
		field := objT.Field(i)
		if !field.IsExported() {
			continue
		}
		key := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			// 取逗号前的部分（去掉 omitempty 等选项）
			if idx := len(tag); idx > 0 {
				for j, c := range tag {
					if c == ',' {
						idx = j
						break
					}
				}
				key = tag[:idx]
			}
		}
		data[key] = objV.Field(i).Interface()
	}
	return data, nil
}

// HttpClient 返回启用了 TLS 证书验证的 HTTP 客户端。
func HttpClient() *http.Client {
	return &http.Client{}
}
