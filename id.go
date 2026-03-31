package gowk

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

func UUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

// GenerateRandomString 生成指定长度的随机字符串。
// crypto/rand 失败时重试，最多 3 次后 panic。
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLen := big.NewInt(int64(len(charset)))
	b := make([]byte, length)
	for i := range b {
		var n *big.Int
		var err error
		for attempt := 0; attempt < 3; attempt++ {
			n, err = rand.Int(rand.Reader, charsetLen)
			if err == nil {
				break
			}
		}
		if err != nil {
			panic(fmt.Sprintf("crypto/rand 不可用: %v", err))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
