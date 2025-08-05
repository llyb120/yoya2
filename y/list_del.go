package y

type delFunc[T any] interface {
	func(item T, index int) bool | func(item T) bool | func(item *T, index int) bool | func(item *T) bool | option
}

func Del[T any, K delFunc[T]](arr []T, fn K, opts ...any) []T {
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
	return nil
}
