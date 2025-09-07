package y

type distinctFunc[T any] interface {
	func(T, int) any | func(*T, int) any | func(T) any | func(*T) any
}

func Distinct[T any](arr []T, fn ...any) []T {
	var mp = make(map[any]bool)
	var result []T
	for i, v := range arr {
		var k any
		if len(fn) > 0 {
			k = doDistinct(fn[0], &arr[i], i)
		} else {
			k = v
		}
		if mp[k] {
			continue
		}
		result = append(result, v)
		mp[k] = true
	}
	return result
}

func doDistinct[T any](fn any, v *T, i int) any {
	switch fn := fn.(type) {
	case func(T, int) any:
		return fn(*v, i)
	case func(*T, int) any:
		return fn(v, i)
	case func(T) any:
		return fn(*v)
	case func(*T) any:
		return fn(v)
	default:
		return *v
	}
}
