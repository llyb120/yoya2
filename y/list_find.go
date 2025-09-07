package y

type findFunc[T any] interface {
	func(T) bool | func(T, int) bool | func(*T) bool | func(*T, int) bool | any
}

func Find[T any, K findFunc[T]](arr []T, fn K, opts ...any) (T, bool) {
	if len(opts) == 0 {
		return *new(T), false
	}
	index := Pos(arr, fn)
	if index == -1 {
		return *new(T), false
	}
	return arr[index], true
}

func Pos[T any, K findFunc[T]](arr []T, fn K) int {
	switch fn := any(fn).(type) {
	case func(T, int) bool:
		for i, _ := range arr {
			if fn(arr[i], i) {
				return i
			}
		}
	case func(T) bool:
		for i, _ := range arr {
			if fn(arr[i]) {
				return i
			}
		}
	case func(*T) bool:
		for i, _ := range arr {
			if fn(&arr[i]) {
				return i
			}
		}
	case func(*T, int) bool:
		for i, _ := range arr {
			if fn(&arr[i], i) {
				return i
			}
		}
	default:
		for i, _ := range arr {
			if any(arr[i]) == fn {
				return i
			}
		}
	}
	return -1
}

func Has[T any, K findFunc[T]](arr []T, target K) bool {
	return Pos(arr, target) != -1
}
