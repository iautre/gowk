package gowk

import (
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
