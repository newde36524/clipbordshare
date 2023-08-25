package invoke

import (
	"sync"
	"sync/atomic"
)

/*
example:
	a, b := AlternateDo()
	go func() {
		for i := 1; i <= 26; i++ {
			a(func() {
				fmt.Println(int(i))
			})
		}
	}()
	go func() {
		for i := byte(65); i <= byte(90); i++ {
			b(func() {
				fmt.Println(string(i))
			})
		}
	}()
**/

func AlternateDo() (a, b func(fn func())) {
	lock := &sync.RWMutex{}
	sendCond := sync.NewCond(lock)
	temp := int32(0)
	f := func(i int32) func(fn func()) {
		return func(fn func()) {
			lock.Lock()
			for atomic.LoadInt32(&temp) == i {
				sendCond.Wait()
			}
			atomic.StoreInt32(&temp, i)
			fn()
			lock.Unlock()
			sendCond.Signal()
		}
	}
	return f(1), f(0)
}
