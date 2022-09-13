package GoroutinePool

import (
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var workerPool sync.Pool

func init() {
	workerPool.New = newWorker
}

type worker struct {
	pool *Pool
}

func newWorker() interface{} {
	return &worker{}
}

// 工作协程循环从taskList中取出task并执行
func (w *worker) run() {
	go func() {
		for {
			var t *task
			w.pool.taskLock.Lock()
			if w.pool.listHead != nil {
				t = w.pool.listHead
				w.pool.listHead = w.pool.listHead.next
				atomic.AddInt32(&w.pool.taskCount, -1)
			}
			if t == nil {
				// 没有任务时则返回
				w.close()
				w.pool.taskLock.Unlock()
				w.Recycle()
				return
			}
			w.pool.taskLock.Unlock()
			func() {
				defer func() {
					if r := recover(); r != nil {
						if w.pool.panicHandler != nil {
							w.pool.panicHandler(t.ctx, r)
						} else {
							msg := fmt.Sprintf("GoroutinePool: panic:%s, %v, %s", w.pool.name, r, debug.Stack())
							fmt.Println(msg)
						}
					}
				}()
				t.f()
			}()
			t.Recycle()
		}
	}()
}

func (w *worker) close() {
	w.pool.decWorkerCount()
}

func (w *worker) SetNil() {
	w.pool = nil
}

func (w *worker) Recycle() {
	w.SetNil()
	workerPool.Put(w)
}
