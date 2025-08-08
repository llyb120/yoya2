package y

import (
	"log"
	"reflect"
	"runtime"
)

func Flex[T any, R any](arr []T, fn func(T, int) R, opts ...option) []R {
	var async, distinct, ignoreNil, ignoreEmpty, isPanic bool
	for _, opt := range opts {
		switch opt {
		case UseAsync:
			async = true
		case UseDistinct:
			distinct = true
		case NotNil:
			ignoreNil = true
		case NotEmpty:
			ignoreEmpty = true
		case UsePanic:
			isPanic = true
		}
	}
	result := make([]R, len(arr))
	if async {
		var wg = WaitGroup{}
		wg.SetLimit(runtime.GOMAXPROCS(0))
		for i, _ := range arr {
			i := i
			wg.goWithPanic(func() error {
				defer func() {
					if r := recover(); r != nil {
						if isPanic {
							panic(r)
						}
						log.Println("panic: ", r)
						wg.Lock()
						defer wg.Unlock()
						result[i] = *new(R)
					}
				}()
				r := fn(arr[i], i)
				wg.Lock()
				defer wg.Unlock()
				result[i] = r
				return nil
			})
		}
		wg.Wait()
	} else {
		// 同步
		for i, _ := range arr {
			r := fn(arr[i], i)
			result[i] = r
		}
	}
	// 如果需要过滤
	if ignoreNil || ignoreEmpty {
		result = Filter(result, func(v R) bool {
			if ignoreNil {
				rv := any(v)
				if rv == nil {
					return false
				}
				val := reflect.ValueOf(rv)
				// 检查是否为指针、切片、映射、通道、函数或接口
				switch val.Kind() {
				case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
					if val.IsNil() {
						return false
					}
				}
			}
			if ignoreEmpty {
				rv := any(v)
				if rv == nil {
					return false
				}
				if isZero(rv) {
					return false
				}
			}
			return true
		})
	}
	if distinct {
		result = Distinct(result, func(v R, i int) any {
			return v
		})
	}
	return result
}

func FlatFlex[T any, R any](arr []T, fn func(T, int) []R, opts ...option) []R {
	var _result = Flex(arr, fn, opts...)
	var result = make([]R, 0)
	for _, v := range _result {
		result = append(result, v...)
	}
	return result
}
