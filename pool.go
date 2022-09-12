package GoroutinePool

import (
	"context"
	"sync"
	"sync/atomic"
)

type IPool interface {
	// Name 返回池子名字
	Name() string
	// SetCap 设置协程池大小
	SetCap(int32)
	// Go 执行任务函数
	Go(func())
	// CtxGo 接收上下文，并且执行任务函数
	CtxGo(context.Context, func())
	// SetPanicHandler 设置panic处理函数
	SetPanicHandler(func(context.Context, interface{}))
	// WorkerCount 返回工作协程数量
	WorkerCount() int32
}

var taskPool sync.Pool

func init() {
	taskPool.New = newTask
}

func newTask() interface{} {
	return &task{}
}

type task struct {
	ctx  context.Context
	f    func()
	next *task
}

func (t *task) Nil() {
	t.ctx = nil
	t.f = nil
	t.next = nil
}

func (t *task) Recycle() {
	t.Nil()
	taskPool.Put(t)
}

type Pool struct {
	// 池子名
	name string
	// 池子容量
	cap int32
	// 配置数据
	config *Config
	// 任务列表
	taskLock  sync.Mutex
	listHead  *task
	listTail  *task
	taskCount int32 // 任务数量
	// 工作人数量
	workerCount int32
	// worker panic时触发
	panicHandler func(context.Context, interface{})
}

func NewPool(name string, cap int32, config *Config) IPool {
	return &Pool{
		name:         name,
		cap:          cap,
		config:       config,
		listHead:     nil,
		listTail:     nil,
		taskLock:     sync.Mutex{},
		taskCount:    0,
		workerCount:  0,
		panicHandler: nil,
	}
}

func (p *Pool) Name() string {
	return p.name
}

func (p *Pool) SetCap(cap int32) {
	atomic.StoreInt32(&p.cap, cap)
}

func (p *Pool) Go(f func()) {
	p.CtxGo(context.Background(), f)
}

func (p *Pool) CtxGo(ctx context.Context, f func()) {
	t := taskPool.Get().(*task)
	t.ctx = ctx
	t.f = f
	p.taskLock.Lock()
	if p.listHead == nil {
		p.listHead = t
		p.listTail = t
	} else {
		p.listTail.next = t
		p.listTail = t
	}
	p.taskLock.Unlock()
	atomic.AddInt32(&p.taskCount, 1)
	// 需要满足两个条件之中的一个：
	// 1.1、任务数量必须大于阈值。
	// 1.2、当前goroutine的数量少于cap容量。
	// 2、goroutine数量为0时。
	if (atomic.LoadInt32(&p.taskCount) >= p.config.ScaleThreshold && p.WorkerCount() < atomic.LoadInt32(&p.cap)) || p.WorkerCount() == 0 {
		p.incWorkerCount()
		w := workerPool.Get().(*worker)
		w.pool = p
		w.run()
	}
}

// SetPanicHandler 设置panic处理函数，当协程panic时，会在recover中恢复并执行回调
func (p *Pool) SetPanicHandler(f func(context.Context, interface{})) {
	p.panicHandler = f
}

func (p *Pool) WorkerCount() int32 {
	return atomic.LoadInt32(&p.workerCount)
}

func (p *Pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCount, 1)
}

func (p *Pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCount, -1)
}
