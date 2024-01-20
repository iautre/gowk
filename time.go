package gowk

import (
	"database/sql/driver"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Time struct {
	time.Time
}

const (
	TimeFormart = "2006-01-02 15:04:05.000"
)

// 实现它的json序列化方法
func (t Time) MarshalJSON() ([]byte, error) {
	var timeStr = ""
	if !t.IsZero() {
		timeStr = t.Format(TimeFormart)
	}
	var stamp = fmt.Sprintf("\"%s\"", timeStr)
	return []byte(stamp), nil
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	timeStr := string(data)
	if timeStr == "\"\"" {
		*t = Time{}
		return
	}
	now, err := time.ParseInLocation("\""+TimeFormart+"\"", timeStr, time.Local)
	*t = Time{now}
	return
}

func Now() *Time {
	return &Time{time.Now()}
}

// Value insert timestamp into mysql need this function.
func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *Time) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = Time{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// 实现 bson 的 序列化方法
func (t *Time) MarshalBSONValue() (bsontype.Type, []byte, error) {
	timestampt := t.Time.Format(TimeFormart)
	retByte := make([]byte, 0)
	retByte = bsoncore.AppendString(retByte, timestampt)
	return bsontype.String, retByte, nil
}

// 实现 bson 的 反序列化方法
func (t *Time) UnmarshalBSONValue(ty bsontype.Type, data []byte) error {
	if ty == bsontype.String {
		if readString, _, ok := bsoncore.ReadString(data); ok {
			now, err := time.ParseInLocation(TimeFormart, readString, time.Local)
			if err != nil {
				return err
			}
			*t = Time{now}
		}
	}
	return nil
}
