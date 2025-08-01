package y

import (
	"reflect"
)

type filterFunc[T any] interface {
	func(T) bool | func(T, int) bool | func(*T) bool | func(*T, int) bool | option
}

func Filter[T any, K filterFunc[T]](arr []T, fn K, opts ...any) []T {
	if len(opts) == 0 {
		return arr
	}
	// 如果第一个是一个方法
	switch fn := any(fn).(type) {
	case func(T) bool:
		return filter0(arr, fn)
	case func(T, int) bool:
		return filter1(arr, fn)
	case func(*T) bool:
		return filter2(arr, fn)
	case func(*T, int) bool:
		return filter3(arr, fn)
	default:
		// 处理opts
		opts = append([]any{fn}, opts...)
	}
	filterOption := struct {
		include     []any
		exclude     []any
		ignoreNil   bool
		ignoreEmpty bool
	}{
		include:     make([]any, 0),
		exclude:     make([]any, 0),
		ignoreNil:   false,
		ignoreEmpty: false,
	}
	var last option
	for _, opt := range opts {
		if realOpt, ok := opt.(option); ok {
			switch realOpt {
			case Is:
				last = realOpt
			case Not:
				last = realOpt
			case NotNil:
				filterOption.ignoreNil = true
			case NotEmpty:
				filterOption.ignoreEmpty = true
			}
		} else {
			switch last {
			case Is:
				filterOption.include = append(filterOption.include, opt)
			case Not:
				filterOption.exclude = append(filterOption.exclude, opt)
			default:
				filterOption.include = append(filterOption.include, opt)
			}
		}
	}
	var result = make([]T, 0, len(arr))
	var shouldContinue bool
	for _, v := range arr {
		if filterOption.ignoreNil && isNil(v) {
			continue
		}
		if filterOption.ignoreEmpty && isZero(v) {
			continue
		}
		shouldContinue = true
		for _, exclude := range filterOption.exclude {
			if exclude == any(v) {
				shouldContinue = false
				break
			}
		}
		if !shouldContinue {
			continue
		}
		if len(filterOption.include) > 0 {
			for _, include := range filterOption.include {
				if include == any(v) {
					result = append(result, v)
					break
				}
			}
		} else {
			result = append(result, v)
		}
	}
	return result
}

func filter0[T any](arr []T, fn func(T) bool) []T {
	var result = make([]T, 0, len(arr))
	for _, v := range arr {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func filter1[T any](arr []T, fn func(T, int) bool) []T {
	var result = make([]T, 0, len(arr))
	for i, v := range arr {
		if fn(v, i) {
			result = append(result, v)
		}
	}
	return result
}

func filter2[T any](arr []T, fn func(*T) bool) []T {
	var result = make([]T, 0, len(arr))
	for i, _ := range arr {
		if fn(&arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

func filter3[T any](arr []T, fn func(*T, int) bool) []T {
	var result = make([]T, 0, len(arr))
	for i, _ := range arr {
		if fn(&arr[i], i) {
			result = append(result, arr[i])
		}
	}
	return result
}

func isNil(v any) bool {
	return v == nil
}

func isZero[T any](v T) bool {
	switch v := any(v).(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case bool:
		return v == false
	default:
		return reflect.ValueOf(v).IsZero()
	}
}
