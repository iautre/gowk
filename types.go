package gowk

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type DateRange [2]string

// Scan 实现 Scanner 接口
func (dr *DateRange) Scan(value interface{}) error {
	// 将数据库中的值转换为 DateRange 类型
	if value == nil {
		return nil
	}
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("DateRange Scan failed: %v is not a string", value)
	}
	endStr := val[len(val)-1 : len(val)]
	// 去除字符串中的括号
	val = val[1 : len(val)-1]
	// 分割字符串为起始和结束日期
	dates := strings.Split(val, ",")
	if len(dates) != 2 {
		return fmt.Errorf("DateRange Scan failed: %s is not a valid date range", val)
	}
	if dates[0] == "" || dates[1] == "" {
		return nil
	}
	// 解析起始和结束日期
	start, err := time.Parse(time.DateOnly, strings.TrimSpace(dates[0]))
	if err != nil {
		return err
	}
	dr[0] = start.Format(time.DateOnly)
	end, err := time.Parse(time.DateOnly, strings.TrimSpace(dates[1]))
	if err != nil {
		return err
	}
	//处理结束日期
	if endStr == ")" {
		end = end.AddDate(0, 0, -1)
	}
	dr[1] = end.Format(time.DateOnly)
	return nil
}

// Value 实现 Valuer 接口
func (dr DateRange) Value() (driver.Value, error) {
	if dr[0] == "" || dr[1] == "" {
		return nil, nil
	}
	// 将 DateRange 类型转换为数据库可以存储的值
	return fmt.Sprintf("[%s,%s]", dr[0], dr[1]), nil
}
