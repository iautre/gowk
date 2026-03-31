package gowk

import (
	"encoding/json"
	"fmt"
)

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
