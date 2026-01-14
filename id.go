package gowk

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iautre/snowflake"
	"github.com/sony/sonyflake"
)

// const timeStr string = "2017-02-27 17:30:20"

var sonyId *sonyflake.Sonyflake = sonyflake.NewSonyflake(sonyflake.Settings{
	StartTime: time.Unix(1655824298, 0),
})

func SonyflakeID() uint64 {
	id, _ := sonyId.NextID()
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

// Helper functions
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to less secure method if crypto/rand fails
			n = big.NewInt(int64(len(charset)))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
