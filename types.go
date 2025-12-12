package gowk

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateRange 实现 PostgreSQL daterange 映射
type DateRange [2]time.Time

// Scan 实现 sql.Scanner 接口
func (dr *DateRange) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan %T into DateRange", src)
	}

	// 解析 PostgreSQL 范围格式: "[2024-01-01,2024-12-31)"
	s = strings.Trim(s, "[]()")
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid daterange format: %s", s)
	}

	start, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
	if err != nil {
		return err
	}

	dr[0] = start
	dr[1] = end
	return nil
}

// Value 实现 driver.Valuer 接口
func (dr DateRange) Value() (driver.Value, error) {
	if dr[0].IsZero() && dr[1].IsZero() {
		return nil, nil
	}
	return fmt.Sprintf("[%s,%s)",
		dr[0].Format("2006-01-02"),
		dr[1].Format("2006-01-02"),
	), nil
}
