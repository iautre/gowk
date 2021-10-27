package gowk

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Time struct {
	time.Time
}

const (
	timeFormart = "2006-01-02 15:04:05"
)

// 实现它的json序列化方法
func (t Time) MarshalJSON() ([]byte, error) {
	var timeStr = ""
	if !t.IsZero() {
		timeStr = t.Format(timeFormart)
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
	now, err := time.ParseInLocation("\""+timeFormart+"\"", timeStr, time.Local)
	*t = Time{now}
	return
}

func Now() Time {
	return Time{time.Now()}
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
