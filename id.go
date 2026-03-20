package gowk

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iautre/snowflake"
	"github.com/sony/sonyflake"
)

var sonyId *sonyflake.Sonyflake = sonyflake.NewSonyflake(sonyflake.Settings{
	StartTime: time.Unix(1655824298, 0),
})

// SonyflakeID 返回唯一 ID。溢出时记录错误日志并 panic，避免返回 0 导致 ID 冲突。
func SonyflakeID() uint64 {
	id, err := sonyId.NextID()
	if err != nil {
		slog.Error("SonyflakeID 生成失败", "err", err)
		panic(fmt.Sprintf("sonyflake: %v", err))
	}
	return id
}

func SnowflakeID() int64 {
	return snowflake.NextId()
}

func UUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func NewAuid() uint {
	return uint(SonyflakeID())
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
