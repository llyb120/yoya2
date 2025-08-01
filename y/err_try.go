package y

import (
	"fmt"
	"runtime"
)

func Try(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			stackLen := runtime.Stack(stack, false)
			err = fmt.Errorf("panic: %v\nstack: %s", r, stack[:stackLen])
		}
	}()
	fn()
	return nil
}

func TryDo[T any](fn func() T) (v T, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			stackLen := runtime.Stack(stack, false)
			err = fmt.Errorf("panic: %v\nstack: %s", r, stack[:stackLen])
		}
	}()
	v = fn()
	return v, nil
}
