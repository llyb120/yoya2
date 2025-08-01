package y

// import (
// 	// "reflect"

// 	"log"
// 	"reflect"
// 	"sync"

// 	"github.com/llyb120/yoya/stlx"
// 	"gvisor.dev/gvisor/pkg/goid"
// 	// "github.com/llyb120/yoya/internal"
// 	// "github.com/llyb120/yoya/stlx"
// 	// "github.com/llyb120/yoya/syncx"
// )

// var zero = &struct {
// 	bool
// 	_ struct{}
// }{}

// type lsxOptionContext struct {
// 	ignoreNil   bool
// 	ignoreEmpty bool
// 	async       bool
// 	distinct    bool
// 	optionIndex []int
// }

// func setIndex(pos int) {
// 	loopIndexMap.Store(goid.Get(), pos)
// }

// func setIndexWithGoid(id int, pos int) {
// 	loopIndexMap.Store(id, pos)
// }

// func GetIndex() int {
// 	id := goid.Get()
// 	if pos, ok := loopIndexMap.Load(id); ok {
// 		return pos.(int)
// 	}
// 	return -1
// }

// var loopIndexMap = sync.Map{}

// func scanOptions(opts []any) *lsxOptionContext {
// 	ctx := &lsxOptionContext{}
// 	for i, opt := range opts {
// 		switch opt {
// 		case NotNil:
// 			ctx.ignoreNil = true
// 		case NotEmpty:
// 			ctx.ignoreEmpty = true
// 		case UseAsync:
// 			ctx.async = true
// 		case UseDistinct:
// 			ctx.distinct = true
// 		}
// 		if _, ok := opt.(option); ok {
// 			ctx.optionIndex = append(ctx.optionIndex, i)
// 		}
// 	}
// 	return ctx
// }

// func Map[T any, R any](arr []T, fn func(T, int) R, opts ...option) []R {
// 	ctx := scanOptions(opts)
// 	result := make([]R, len(arr))
// 	if ctx.async {
// 		var wg = WaitGroup{}
// 		for i, v := range arr {
// 			i := i
// 			v := v
// 			wg.goWithPanic(func() error {
// 				defer func() {
// 					if r := recover(); r != nil {
// 						log.Println("panic: ", r)
// 						wg.Lock()
// 						defer wg.Unlock()
// 						result[i] = *new(R)
// 					}
// 				}()
// 				r := fn(v, i)
// 				wg.Lock()
// 				defer wg.Unlock()
// 				result[i] = r
// 				return nil
// 			})
// 		}
// 		wg.Wait()
// 	} else {
// 		// 同步
// 		for i, v := range arr {
// 			r := fn(v, i)
// 			result[i] = r
// 		}
// 	}
// 	// 如果需要过滤
// 	if ctx.ignoreNil || ctx.ignoreEmpty {
// 		Filter(&result, func(v R) bool {
// 			if ctx.ignoreNil {
// 				rv := any(v)
// 				if rv == nil {
// 					return false
// 				}
// 				val := reflect.ValueOf(rv)
// 				// 检查是否为指针、切片、映射、通道、函数或接口
// 				switch val.Kind() {
// 				case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
// 					if val.IsNil() {
// 						return false
// 					}
// 				}
// 			}
// 			if ctx.ignoreEmpty {
// 				rv := any(v)
// 				if rv == nil {
// 					return false
// 				}
// 				if isZero(rv) {
// 					return false
// 				}
// 			}
// 			return true
// 		})
// 	}
// 	if ctx.distinct {
// 		Distinct(&result, func(v R, i int) any {
// 			return v
// 		})
// 	}
// 	return result
// }

// func FlatMap[T any, R any](arr []T, fn func(T, int) []R, opts ...lsxOption) []R {
// 	var _result = Map(arr, fn, opts...)
// 	var result = make([]R, 0)
// 	for _, v := range _result {
// 		result = append(result, v...)
// 	}
// 	return result
// }

// func Find[T any](arr []T, fn any) (T, bool) {
// 	index := Pos(arr, fn)
// 	if index == -1 {
// 		return *new(T), false
// 	}
// 	return arr[index], true
// }

// func Pos[T any](arr []T, fn any) int {
// 	switch fn := fn.(type) {
// 	case func(T, int) bool:
// 		for i, v := range arr {
// 			if fn(v, i) {
// 				return i
// 			}
// 		}
// 	case func(T) bool:
// 		for i, v := range arr {
// 			if fn(v) {
// 				return i
// 			}
// 		}
// 	default:
// 		for i, v := range arr {
// 			if any(v) == fn {
// 				return i
// 			}
// 		}
// 	}
// 	return -1
// }

// func Has[T any](arr []T, target any) bool {
// 	switch target := target.(type) {
// 	case func(T) bool:
// 		for _, v := range arr {
// 			if target(v) {
// 				return true
// 			}
// 		}
// 		return false

// 	case func(T, int) bool:
// 		for i, v := range arr {
// 			if target(v, i) {
// 				return true
// 			}
// 		}
// 		return false
// 	default:
// 		for _, v := range arr {
// 			if any(v) == target {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// }

// func Del[T comparable](arr *[]T, pos int) {
// 	if pos < 0 || pos >= len(*arr) {
// 		return
// 	}
// 	*arr = append((*arr)[:pos], (*arr)[pos+1:]...)
// }

// func Reduce[T any, R any](arr []T, fn func(R, T) R, initial R) R {
// 	var source []T
// 	result := initial
// 	source = arr
// 	for _, v := range source {
// 		result = fn(result, v)
// 	}
// 	return result
// }

// func For[T any](arr []T, fn func(T, int) bool) {
// 	var source []T
// 	source = arr
// 	for i, v := range source {
// 		c := fn(v, i)
// 		if !c {
// 			return
// 		}
// 	}
// }

// func Distinct[T any, A array[T]](arr A, fn func(T, int) any) {
// 	var mp = make(map[any]bool)
// 	var result []T
// 	for i, v := range *arr {
// 		var k any
// 		if len(fn) > 0 {
// 			k = fn[0](v, i)
// 		} else {
// 			k = v
// 		}
// 		if mp[k] {
// 			continue
// 		}
// 		result = append(result, v)
// 		mp[k] = true
// 	}
// 	*arr = result
// }

// func Sort[T any](arr *[]T, less func(T, T) bool) {
// 	var cp []T
// 	cp = make([]T, len(*arr))
// 	copy(cp, *arr)
// 	cp = timSort(cp, less)
// 	*arr = cp
// }

// func Keys[K comparable, V any](mp map[K]V) []K {
// 	result := make([]K, 0, len(mp))
// 	for k := range mp {
// 		result = append(result, k)
// 	}
// 	return result
// }

// func Vals[K comparable, V any](mp map[K]V) []V {
// 	result := make([]V, 0, len(mp))
// 	for _, v := range mp {
// 		result = append(result, v)
// 	}
// 	return result
// }

// // func Mock[K any, T any](arr *[]K, fn func(*[]T)) error {
// // 	var mock []T
// // 	err := internal.Cast(&mock, arr)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	fn(&mock)
// // 	// 还原
// // 	var result []K
// // 	err = internal.Cast(&result, mock)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	*arr = result
// // 	return nil
// // }

// func Group[T any](arr []T, fn func(T, int) any) [][]T {
// 	var result = stlx.NewMultiMap[any, T]()
// 	for i, v := range arr {
// 		k := fn(v, i)
// 		if k == nil {
// 			continue
// 		}
// 		result.Set(k, v)
// 	}
// 	return result.Vals()
// }

// func Chunk[T any](arr []T, size int) [][]T {
// 	var result = make([][]T, 0, len(arr)/size+1)
// 	for i := 0; i < len(arr); i += size {
// 		end := i + size
// 		if end > len(arr) {
// 			end = len(arr)
// 		}
// 		result = append(result, arr[i:end])
// 	}
// 	return result
// }

// func GroupMap[K comparable, V any](arr []V, fn func(V, int) K) map[K][]V {
// 	var result = make(map[K][]V)
// 	for i, v := range arr {
// 		k := fn(v, i)
// 		result[k] = append(result[k], v)
// 	}
// 	return result
// }

// func ToMap[K comparable, V any](arr []V, fn func(V, int) K) map[K]V {
// 	var result = make(map[K]V)
// 	for i, v := range arr {
// 		k := fn(v, i)
// 		result[k] = v
// 	}
// 	return result
// }

// func timSort[T any](arr []T, less func(T, T) bool) []T {
// 	if len(arr) <= 1 {
// 		return arr
// 	}

// 	const minRun = 32
// 	const maxRun = 64

// 	// 计算最小运行长度
// 	runLength := minRun
// 	if len(arr) < minRun {
// 		runLength = len(arr)
// 	}

// 	// 插入排序辅助函数
// 	insertionSort := func(arr []T, left, right int) {
// 		for i := left + 1; i <= right; i++ {
// 			key := arr[i]
// 			j := i - 1
// 			for j >= left && less(key, arr[j]) {
// 				arr[j+1] = arr[j]
// 				j--
// 			}
// 			arr[j+1] = key
// 		}
// 	}

// 	// 合并两个有序子数组
// 	merge := func(arr []T, l, m, r int) {
// 		left := make([]T, m-l+1)
// 		right := make([]T, r-m)
// 		copy(left, arr[l:m+1])
// 		copy(right, arr[m+1:r+1])

// 		i, j, k := 0, 0, l
// 		for i < len(left) && j < len(right) {
// 			if !less(right[j], left[i]) {
// 				arr[k] = left[i]
// 				i++
// 			} else {
// 				arr[k] = right[j]
// 				j++
// 			}
// 			k++
// 		}

// 		for i < len(left) {
// 			arr[k] = left[i]
// 			i++
// 			k++
// 		}

// 		for j < len(right) {
// 			arr[k] = right[j]
// 			j++
// 			k++
// 		}
// 	}

// 	// 对小块使用插入排序
// 	for i := 0; i < len(arr); i += runLength {
// 		end := i + runLength - 1
// 		if end >= len(arr) {
// 			end = len(arr) - 1
// 		}
// 		insertionSort(arr, i, end)
// 	}

// 	// 合并排序后的小块
// 	for size := runLength; size < len(arr); size *= 2 {
// 		for start := 0; start < len(arr); start += size * 2 {
// 			mid := start + size - 1
// 			end := start + size*2 - 1
// 			if mid >= len(arr) {
// 				mid = len(arr) - 1
// 			}
// 			if end >= len(arr) {
// 				end = len(arr) - 1
// 			}
// 			if mid > start {
// 				merge(arr, start, mid, end)
// 			}
// 		}
// 	}

// 	return arr
// }

// func isZero[T any](v T) bool {
// 	switch v := any(v).(type) {
// 	case string:
// 		return v == ""
// 	case int, int8, int16, int32, int64:
// 		return v == 0
// 	case uint, uint8, uint16, uint32, uint64:
// 		return v == 0
// 	case float32, float64:
// 		return v == 0
// 	case bool:
// 		return v == false
// 	default:
// 		return reflect.ValueOf(v).IsZero()
// 	}
// }
