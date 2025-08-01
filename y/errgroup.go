package y

import (
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

type WaitGroup struct {
	errgroup.Group
	sync.RWMutex
}

func (g *WaitGroup) Go(f func() error) {
	g.Group.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				// 打印调用栈
				buf := make([]byte, 1024)
				n := runtime.Stack(buf, false)
				err = fmt.Errorf("panic: %v\n%s", r, buf[:n])
			}
		}()
		return f()
	})
}

func (g *WaitGroup) goWithPanic(f func() error) {
	g.Group.Go(f)
}
