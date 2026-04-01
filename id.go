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

// UUIDV7 生成 RFC 9562 UUID v7（时间排序、可索引）。失败时返回错误（熵源异常等）。
func UUIDV7() (uuid.UUID, error) {
	return uuid.NewV7()
}

// MustUUIDV7 生成 UUID v7，失败时 panic（与 GenerateRandomString 策略一致，仅用于确信可用的环境）。
func MustUUIDV7() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("gowk: uuid v7: %v", err))
	}
	return u
}

// UUIDV7String 返回标准带连字符的 UUID v7 字符串。
func UUIDV7String() string {
	return MustUUIDV7().String()
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
