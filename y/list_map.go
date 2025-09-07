package y

import (
	"log"
	"reflect"
	"runtime"
)

type flexOption struct {
	ignoreNil   bool
	ignoreEmpty bool
	isPanic     bool
	async       bool
	distinct    bool
	isFlatFlex  bool
}

func Flex[T any, R any](arr []T, fn func(T, int) R, opts ...option) []R {
	var flexOption flexOption
	makeFlexOption(&flexOption, opts...)
	result := make([]R, len(arr))
	if flexOption.async {
		var wg = WaitGroup{}
		wg.SetLimit(runtime.GOMAXPROCS(0))
		for i, _ := range arr {
			i := i
			wg.goWithPanic(func() error {
				defer func() {
					if r := recover(); r != nil {
						if flexOption.isPanic {
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

	if flexOption.isFlatFlex {
		return result
	}
	return applyFlexOption(&flexOption, result)
}

func FlatFlex[T any, R any](arr []T, fn func(T, int) []R, opts ...option) []R {
	opts = append(opts, isFlatFlex)
	var flexOption flexOption
	makeFlexOption(&flexOption, opts...)
	var _result = Flex(arr, fn, opts...)
	var result = make([]R, 0)
	for _, v := range _result {
		result = append(result, v...)
	}
	return applyFlexOption(&flexOption, result)
}

func makeFlexOption(flexOption *flexOption, opts ...option) {
	for _, opt := range opts {
		switch opt {
		case UseAsync:
			flexOption.async = true
		case UseDistinct:
			flexOption.distinct = true
		case NotNil:
			flexOption.ignoreNil = true
		case NotEmpty:
			flexOption.ignoreEmpty = true
		case UsePanic:
			flexOption.isPanic = true
		case isFlatFlex:
			flexOption.isFlatFlex = true
		}
	}
}

func applyFlexOption[R any](flexOption *flexOption, result []R) []R {
	// 如果需要过滤
	if flexOption.ignoreNil || flexOption.ignoreEmpty {
		result = Filter(result, func(v R) bool {
			if flexOption.ignoreNil {
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
			if flexOption.ignoreEmpty {
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
	if flexOption.distinct {
		result = Distinct(result, func(v R, i int) any {
			return v
		})
	}
	return result
}
