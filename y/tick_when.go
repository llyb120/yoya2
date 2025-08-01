package y

import (
	"time"
)

const (
	EQ  = "=="
	NE  = "!="
	GT  = ">"
	GE  = ">="
	LT  = "<"
	LE  = "<="
	MEQ = "~="
	MGT = "~>"
	MGE = "~>="
	MLT = "~<"
	MLE = "~<="
)

func When(args ...any) bool {
	if len(args) < 3 || len(args)%2 == 0 {
		return false
	}
	if len(args) > 3 {
		// 需要满足 123 345 这种多重关系
		var flag bool = true
		var useFlag = false
		for i := 0; i < len(args); i += 2 {
			if i+1 >= len(args) || i+2 >= len(args) {
				break
			}
			operator, ok := args[i+1].(string)
			if !ok {
				return false
			}
			left := args[i]
			right := args[i+2]
			flag = flag && When(left, operator, right)
			useFlag = true
		}
		return flag && useFlag
	}
	// 这里只有3的情况了
	left := args[0]
	operator, ok := args[1].(string)
	if !ok {
		return false
	}
	right := args[2]
	cache := compareHolder.Get()
	str, isStr := any(left).(string)
	var leftTime time.Time
	// 获取左值
	if isStr {
		leftTime = cache.GetOrSetFunc(str, func() time.Time {
			t, err := Guess(str)
			if err != nil {
				return time.Time{}
			}
			return t
		})
	} else {
		leftTime = any(left).(time.Time)
	}
	// 获取右值
	str, isStr = any(right).(string)
	var rightTime time.Time
	if isStr {
		rightTime = cache.GetOrSetFunc(str, func() time.Time {
			t, err := Guess(str)
			if err != nil {
				return time.Time{}
			}
			return t
		})
	} else {
		rightTime = any(right).(time.Time)
	}
	return tickCompare(leftTime, operator, rightTime)
}

func tickCompare(leftTime time.Time, operator string, rightTime time.Time) bool {
	switch operator {
	case GT:
		return leftTime.After(rightTime)
	case GE:
		return leftTime.After(rightTime) || leftTime.Equal(rightTime)
	case LT:
		return leftTime.Before(rightTime)
	case LE:
		return leftTime.Before(rightTime) || leftTime.Equal(rightTime)
	case EQ:
		return leftTime.Equal(rightTime)
	case NE:
		return !leftTime.Equal(rightTime)
	case MEQ:
		return leftTime.Year() == rightTime.Year() && leftTime.Month() == rightTime.Month() && leftTime.Day() == rightTime.Day()
	case MGT:
		return leftTime.Year() > rightTime.Year() ||
			(leftTime.Year() == rightTime.Year() && leftTime.Month() > rightTime.Month()) ||
			(leftTime.Year() == rightTime.Year() && leftTime.Month() == rightTime.Month() && leftTime.Day() > rightTime.Day())
	case MLT:
		return leftTime.Year() < rightTime.Year() ||
			(leftTime.Year() == rightTime.Year() && leftTime.Month() < rightTime.Month()) ||
			(leftTime.Year() == rightTime.Year() && leftTime.Month() == rightTime.Month() && leftTime.Day() < rightTime.Day())
	case MGE:
		return tickCompare(leftTime, MGT, rightTime) || tickCompare(leftTime, MEQ, rightTime)
	case MLE:
		return tickCompare(leftTime, MLT, rightTime) || tickCompare(leftTime, MEQ, rightTime)
	default:
		return false
	}
}

var compareHolder Holder[*BaseCache[string, time.Time]]

func init() {
	compareHolder.InitFunc = func() *BaseCache[string, time.Time] {
		return NewBaseCache[string, time.Time](CacheOption{
			MaxMemory: "10m",
			TTL:       10 * time.Minute,
		})
	}
}
