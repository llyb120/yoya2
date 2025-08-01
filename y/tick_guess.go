package y

import (
	"fmt"
	"strings"
	"time"
)

func Guess(dateStr string) (time.Time, error) {
	t, _, err := guess(dateStr)
	return t, err
}

// Guess 函数尝试按优先级从上到下解析字符串时间
func guess(dateStr string) (time.Time, string, error) {
	// 去除可能的空白字符
	dateStr = strings.TrimSpace(dateStr)

	if dateStr == "" {
		return time.Time{}, "", fmt.Errorf("日期字符串为空")
	}

	// 根据长度尝试解析
	var formats []string
	switch len(dateStr) {
	case 4:
		formats = []string{
			"2006", // 仅年份
		}
	case 10:
		formats = []string{
			"2006-01-02", // 标准日期
			"2006/01/02", // 斜杠分隔日期
			"01/02/2006", // 美式日期
		}

	case 16:
		formats = []string{
			"2006-01-02 15:04",
			"2006/01/02 15:04",
		}

	case 19:
		formats = []string{
			"2006-01-02 15:04:05",
			"2006/01/02 15:04:05",
			"2006-01-02T15:04:05",
		}

	case 20:
		formats = []string{
			time.RFC3339, // ISO8601带时区
		}

	case 24, 25:
		formats = []string{
			time.RFC3339, // 带时区的ISO8601
			time.RFC1123Z,
		}

	case 14:
		formats = []string{
			"20060102150405", // 紧凑格式
		}

	case 8:
		formats = []string{
			"20060102", // 紧凑日期
		}

	case 28, 29, 30:
		formats = []string{
			time.RFC3339Nano,
		}

	case 22, 23:
		formats = []string{
			time.RFC850,
			time.RFC1123,
		}

	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, format, nil
		}
	}
	return time.Time{}, "", fmt.Errorf("日期格式错误: %s", dateStr)
}
