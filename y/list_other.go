package y

func Reduce[T any, R any](arr []T, fn func(R, T) R, initial R) R {
	var source []T
	result := initial
	source = arr
	for _, v := range source {
		result = fn(result, v)
	}
	return result
}
