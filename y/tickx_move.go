package y

import "time"

type moveType string
type YmdUnit uint64

const (
	Second time.Duration = time.Second
	Minute time.Duration = time.Minute
	Hour   time.Duration = time.Hour
	Day    YmdUnit       = 10000
	Week   YmdUnit       = 70000
	Month  YmdUnit       = 10000 * 10000
	Year   YmdUnit       = 10000 * 10000 * 10000

	FirstDayOfMonth  moveType = "FirstDayOfMonth"
	LastDayOfMonth   moveType = "LastDayOfMonth"
	FirstDayOfYear   moveType = "FirstDayOfYear"
	LastDayOfYear    moveType = "LastDayOfYear"
	FirstDayOfWeek   moveType = "FirstDayOfWeek"
	LastDayOfWeek    moveType = "LastDayOfWeek"
	FirstDayOfCNWeek moveType = "FirstDayOfCNWeek"
	LastDayOfCNWeek  moveType = "LastDayOfCNWeek"
)

func Move[T string | *string | time.Time | *time.Time](date T, movements ...any) T {
	if len(movements) == 0 {
		return date
	}
	// 使用Guess函数尝试解析日期
	var flag bool = true
	var d string
	var isString bool
	var isPointer bool
	switch s := any(date).(type) {
	case string:
		d = s
		isString = true
	case *string:
		if s == nil {
			return date
		}
		d = *s
		isString = true
		isPointer = true
	case *time.Time:
		if s == nil {
			return date
		}
		isPointer = true
	}

	var t time.Time
	var err error
	var format string

	if isString {
		t, format, err = guess(d)
		if err != nil {
			var zero T
			return zero
		}
	} else {
		if isPointer {
			t = *any(date).(*time.Time)
		} else {
			t = any(date).(time.Time)
		}
	}

	// 按顺序处理每个movement，而不是累加后一次应用
	for _, m := range movements {
		switch m := any(m).(type) {
		case YmdUnit:
			years := int(m / Year)
			months := int((m % Year) / Month)
			days := int((m % Month) / Day)

			if years != 0 || months != 0 || days != 0 {
				t = adjustMonthBoundary(t, years, months, days)
			}
		case time.Duration:
			t = t.Add(m)
		case moveType:
			switch m {
			case FirstDayOfMonth:
				t = time.Date(t.Year(), t.Month(), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
			case LastDayOfMonth:
				t = t.AddDate(0, 1, -t.Day())
			case FirstDayOfYear:
				t = time.Date(t.Year(), 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
			case LastDayOfYear:
				t = time.Date(t.Year(), 12, 31, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
			case FirstDayOfWeek:
				t = t.AddDate(0, 0, -int(t.Weekday()))
			case FirstDayOfCNWeek:
				// 对于中国周（从周一开始），需要特殊处理
				// 如果是周日(0)，需要回退6天；否则回退到周一
				offset := int(t.Weekday())
				if offset == 0 {
					offset = 6
				} else {
					offset -= 1
				}
				t = t.AddDate(0, 0, -offset)
			case LastDayOfWeek:
				t = t.AddDate(0, 0, 6-int(t.Weekday()))
			case LastDayOfCNWeek:
				// 对于中国周（从周一开始），需要特殊处理
				// 如果是周日(0)，需要回退6天；否则回退到周一
				offset := int(t.Weekday())
				if offset == 0 {
					offset = 6
				} else {
					offset -= 1
				}
				t = t.AddDate(0, 0, 6-int(t.Weekday()))
			}
		case bool:
			flag = flag && m
			if !flag {
				return date
			}
		}
	}

	if !flag {
		return date
	}

	if isString {
		// 根据输入日期格式决定输出格式
		res := t.Format(format)
		if isPointer {
			*(any(date).(*string)) = res
			return date
		}
		return any(res).(T)
	} else {
		if isPointer {
			*(any(date).(*time.Time)) = t
			return date
		}
		return any(t).(T)
	}
}

// 获取月份的最后一天
func lastDayOfMonth(year int, month time.Month) int {
	// 获取下个月的第一天，然后减去一天
	firstDayOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDayOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}

// 处理月份边界问题
func adjustMonthBoundary(t time.Time, years, months, days int) time.Time {
	// 记录原始日期的信息
	originalYear := t.Year()
	originalMonth := t.Month()
	originalDay := t.Day()

	// 检查原始日期是否是月末
	lastDayOfOriginalMonth := lastDayOfMonth(originalYear, originalMonth)
	isLastDay := originalDay == lastDayOfOriginalMonth

	// 计算目标年月
	targetYear := originalYear + years
	targetMonth := originalMonth + time.Month(months)

	// 调整月份溢出（例如13月变成下一年的1月）
	for targetMonth > 12 {
		targetYear++
		targetMonth -= 12
	}
	for targetMonth < 1 {
		targetYear--
		targetMonth += 12
	}

	// 确定目标日
	var targetDay int
	if isLastDay {
		// 如果原始日期是月末，则目标日期也应该是月末
		targetDay = lastDayOfMonth(targetYear, targetMonth)
	} else {
		// 如果不是月末，则尝试保持原始日期的日
		lastDayOfTargetMonth := lastDayOfMonth(targetYear, targetMonth)
		if originalDay > lastDayOfTargetMonth {
			// 如果原始日期的日大于目标月份的最后一天，则使用目标月份的最后一天
			targetDay = lastDayOfTargetMonth
		} else {
			targetDay = originalDay
		}
	}

	// 创建新的日期
	newTime := time.Date(targetYear, targetMonth, targetDay,
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	// 再调整天数
	if days != 0 {
		newTime = newTime.AddDate(0, 0, days)
	}

	return newTime
}
